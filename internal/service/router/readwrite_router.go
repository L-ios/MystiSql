package router

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"MystiSql/internal/service/loadbalancer"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

type SQLCategory string

const (
	SQLCategoryRead  SQLCategory = "read"
	SQLCategoryWrite SQLCategory = "write"
)

type ReadWriteRouter struct {
	mu             sync.RWMutex
	topology       map[string]*types.MasterSlaveTopology
	loadBalancer   loadbalancer.LoadBalancer
	delayChecker   DelayChecker
	logger         *zap.Logger
	delayThreshold time.Duration
	readAfterWrite string
}

type DelayChecker interface {
	GetDelay(instanceName string) (time.Duration, error)
}

type RouteResult struct {
	TargetInstance string             `json:"targetInstance"`
	Role           types.InstanceRole `json:"role"`
	Category       SQLCategory        `json:"category"`
	Reason         string             `json:"reason"`
}

func NewReadWriteRouter(lb loadbalancer.LoadBalancer, logger *zap.Logger) *ReadWriteRouter {
	return &ReadWriteRouter{
		topology:       make(map[string]*types.MasterSlaveTopology),
		loadBalancer:   lb,
		logger:         logger,
		delayThreshold: time.Second,
		readAfterWrite: "default",
	}
}

func (r *ReadWriteRouter) SetTopology(topology map[string]*types.MasterSlaveTopology) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.topology = topology
}

func (r *ReadWriteRouter) SetDelayThreshold(threshold time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.delayThreshold = threshold
}

func (r *ReadWriteRouter) SetReadAfterWrite(strategy string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.readAfterWrite = strategy
}

func (r *ReadWriteRouter) SetDelayChecker(checker DelayChecker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.delayChecker = checker
}

func (r *ReadWriteRouter) Route(ctx context.Context, instanceName string, sql string, inTransaction bool) (*RouteResult, error) {
	category := r.categorizeSQL(sql)

	if inTransaction {
		return r.routeToMaster(instanceName, "in_transaction")
	}

	if r.isSelectForUpdate(sql) {
		return r.routeToMaster(instanceName, "select_for_update")
	}

	if r.isSpecialFunction(sql) {
		return r.routeToMaster(instanceName, "special_function")
	}

	if category == SQLCategoryWrite {
		return r.routeToMaster(instanceName, "write_operation")
	}

	return r.routeRead(instanceName)
}

func (r *ReadWriteRouter) categorizeSQL(sql string) SQLCategory {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	fields := strings.Fields(upper)

	if len(fields) == 0 {
		return SQLCategoryRead
	}

	firstWord := fields[0]

	switch firstWord {
	case "SELECT", "SHOW", "EXPLAIN", "DESC", "DESCRIBE":
		return SQLCategoryRead
	case "INSERT", "UPDATE", "DELETE", "REPLACE", "CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME":
		return SQLCategoryWrite
	default:
		return SQLCategoryRead
	}
}

func (r *ReadWriteRouter) isSelectForUpdate(sql string) bool {
	re := regexp.MustCompile(`(?i)\bFOR\s+UPDATE\b`)
	return re.MatchString(sql)
}

func (r *ReadWriteRouter) isSpecialFunction(sql string) bool {
	upper := strings.ToUpper(sql)
	return strings.Contains(upper, "LAST_INSERT_ID()") ||
		strings.Contains(upper, "@@IDENTITY")
}

func (r *ReadWriteRouter) routeToMaster(instanceName, reason string) (*RouteResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	topology, exists := r.topology[instanceName]
	if !exists {
		return &RouteResult{
			TargetInstance: instanceName,
			Role:           types.InstanceRoleMaster,
			Category:       SQLCategoryWrite,
			Reason:         reason,
		}, nil
	}

	if topology.Master == nil {
		return nil, fmt.Errorf("no master available for topology %s", instanceName)
	}

	return &RouteResult{
		TargetInstance: topology.Master.Name,
		Role:           types.InstanceRoleMaster,
		Category:       SQLCategoryWrite,
		Reason:         reason,
	}, nil
}

func (r *ReadWriteRouter) routeRead(instanceName string) (*RouteResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	topology, exists := r.topology[instanceName]
	if !exists {
		return &RouteResult{
			TargetInstance: instanceName,
			Role:           types.InstanceRoleMaster,
			Category:       SQLCategoryRead,
			Reason:         "no_topology",
		}, nil
	}

	availableSlaves := topology.GetAvailableSlaves()
	if len(availableSlaves) == 0 {
		return &RouteResult{
			TargetInstance: topology.Master.Name,
			Role:           types.InstanceRoleMaster,
			Category:       SQLCategoryRead,
			Reason:         "no_available_slaves",
		}, nil
	}

	if r.delayChecker != nil {
		for _, slave := range availableSlaves {
			delay, err := r.delayChecker.GetDelay(slave.Name)
			if err != nil {
				continue
			}
			if delay > r.delayThreshold {
				r.logger.Warn("Slave delay exceeds threshold, routing to master",
					zap.String("slave", slave.Name),
					zap.Duration("delay", delay),
					zap.Duration("threshold", r.delayThreshold))
				return &RouteResult{
					TargetInstance: topology.Master.Name,
					Role:           types.InstanceRoleMaster,
					Category:       SQLCategoryRead,
					Reason:         "slave_delay_exceeded",
				}, nil
			}
		}
	}

	selected, err := r.loadBalancer.Select(availableSlaves)
	if err != nil {
		return &RouteResult{
			TargetInstance: topology.Master.Name,
			Role:           types.InstanceRoleMaster,
			Category:       SQLCategoryRead,
			Reason:         "lb_select_failed",
		}, nil
	}

	return &RouteResult{
		TargetInstance: selected.Name,
		Role:           types.InstanceRoleSlave,
		Category:       SQLCategoryRead,
		Reason:         "load_balanced",
	}, nil
}

func (r *ReadWriteRouter) GetTopologyStatus() map[string]*TopologyStatusInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := make(map[string]*TopologyStatusInfo)
	for name, topo := range r.topology {
		availableSlaves := topo.GetAvailableSlaves()
		masterName := ""
		if topo.Master != nil {
			masterName = topo.Master.Name
		}
		status[name] = &TopologyStatusInfo{
			Name:            name,
			MasterName:      masterName,
			MasterStatus:    string(topo.Master.Status),
			TotalSlaves:     len(topo.Slaves),
			AvailableSlaves: len(availableSlaves),
			TopologyStatus:  string(topo.Status),
		}
	}
	return status
}

type TopologyStatusInfo struct {
	Name            string `json:"name"`
	MasterName      string `json:"masterName"`
	MasterStatus    string `json:"masterStatus"`
	TotalSlaves     int    `json:"totalSlaves"`
	AvailableSlaves int    `json:"availableSlaves"`
	TopologyStatus  string `json:"topologyStatus"`
}
