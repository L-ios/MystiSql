package connection

import (
	"testing"

	"MystiSql/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriverRegistry_RegisterDriver(t *testing.T) {
	r := &DriverRegistry{
		drivers: make(map[types.DatabaseType]ConnectionFactory),
	}

	factory := &mockFactory{}

	err := r.RegisterDriver(types.DatabaseTypeMySQL, factory)
	assert.NoError(t, err)

	err = r.RegisterDriver(types.DatabaseTypeMySQL, factory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestDriverRegistry_RegisterDriver_NilFactory(t *testing.T) {
	r := &DriverRegistry{
		drivers: make(map[types.DatabaseType]ConnectionFactory),
	}

	err := r.RegisterDriver(types.DatabaseTypeMySQL, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestDriverRegistry_GetFactory(t *testing.T) {
	r := &DriverRegistry{
		drivers: make(map[types.DatabaseType]ConnectionFactory),
	}

	factory := &mockFactory{}
	err := r.RegisterDriver(types.DatabaseTypeMySQL, factory)
	require.NoError(t, err)

	got, err := r.GetFactory(types.DatabaseTypeMySQL)
	assert.NoError(t, err)
	assert.Equal(t, factory, got)

	got, err = r.GetFactory(types.DatabaseTypePostgreSQL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, got)
}

func TestDriverRegistry_ListDrivers(t *testing.T) {
	r := &DriverRegistry{
		drivers: make(map[types.DatabaseType]ConnectionFactory),
	}

	drivers := r.ListDrivers()
	assert.Empty(t, drivers)

	factory := &mockFactory{}
	_ = r.RegisterDriver(types.DatabaseTypeMySQL, factory)
	_ = r.RegisterDriver(types.DatabaseTypePostgreSQL, factory)

	drivers = r.ListDrivers()
	assert.Len(t, drivers, 2)
	assert.Contains(t, drivers, types.DatabaseTypeMySQL)
	assert.Contains(t, drivers, types.DatabaseTypePostgreSQL)
}

func TestDriverRegistry_IsDriverRegistered(t *testing.T) {
	r := &DriverRegistry{
		drivers: make(map[types.DatabaseType]ConnectionFactory),
	}

	assert.False(t, r.IsDriverRegistered(types.DatabaseTypeMySQL))

	factory := &mockFactory{}
	_ = r.RegisterDriver(types.DatabaseTypeMySQL, factory)

	assert.True(t, r.IsDriverRegistered(types.DatabaseTypeMySQL))
	assert.False(t, r.IsDriverRegistered(types.DatabaseTypePostgreSQL))
}

func TestDriverRegistry_UnregisterDriver(t *testing.T) {
	r := &DriverRegistry{
		drivers: make(map[types.DatabaseType]ConnectionFactory),
	}

	factory := &mockFactory{}
	_ = r.RegisterDriver(types.DatabaseTypeMySQL, factory)
	assert.True(t, r.IsDriverRegistered(types.DatabaseTypeMySQL))

	r.UnregisterDriver(types.DatabaseTypeMySQL)
	assert.False(t, r.IsDriverRegistered(types.DatabaseTypeMySQL))
}

func TestDriverRegistry_Clear(t *testing.T) {
	r := &DriverRegistry{
		drivers: make(map[types.DatabaseType]ConnectionFactory),
	}

	factory := &mockFactory{}
	_ = r.RegisterDriver(types.DatabaseTypeMySQL, factory)
	_ = r.RegisterDriver(types.DatabaseTypePostgreSQL, factory)

	r.Clear()
	assert.Empty(t, r.ListDrivers())
}

func TestGetRegistry_Singleton(t *testing.T) {
	r1 := GetRegistry()
	r2 := GetRegistry()
	assert.Same(t, r1, r2)
}

type mockFactory struct{}

func (f *mockFactory) CreateConnection(instance *types.DatabaseInstance) (Connection, error) {
	return nil, nil
}
