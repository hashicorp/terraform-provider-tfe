package provider

var _ capabilitiesResolver = &staticCapabilityResolver{}

// A mock capability resolver used for testing to set specific capabilities
type staticCapabilityResolver struct {
	isCloud bool
	tfeVer  string
}

func (r *staticCapabilityResolver) IsCloud() bool {
	return r.isCloud
}

func (r *staticCapabilityResolver) RemoteTFEVersion() string {
	return r.tfeVer
}

func (r *staticCapabilityResolver) SetIsCloud(val bool) {
	r.isCloud = val
}

func (r *staticCapabilityResolver) SetRemoteTFEVersion(val string) {
	r.tfeVer = val
}
