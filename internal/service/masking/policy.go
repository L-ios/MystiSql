package masking

type MaskPolicy struct {
	MaskPhone    bool
	MaskEmail    bool
	MaskIDCard   bool
	MaskBankCard bool
}

type PolicyStore struct {
	Policies map[string]MaskPolicy
}

func NewPolicyStore() *PolicyStore {
	return &PolicyStore{Policies: make(map[string]MaskPolicy)}
}

func (ps *PolicyStore) Get(role string) MaskPolicy {
	if ps == nil {
		return MaskPolicy{}
	}
	if p, ok := ps.Policies[role]; ok {
		return p
	}
	return MaskPolicy{}
}

func (ps *PolicyStore) Set(role string, policy MaskPolicy) {
	if ps.Policies == nil {
		ps.Policies = make(map[string]MaskPolicy)
	}
	ps.Policies[role] = policy
}
