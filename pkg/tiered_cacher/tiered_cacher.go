package tiered_cacher

type TieredCacher struct {
}

func NewTieredCacher(storage interface{}) *TieredCacher {
	tieredCacher := &TieredCacher{}
	return tieredCacher
}
