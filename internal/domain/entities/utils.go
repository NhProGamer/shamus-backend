package entities

func SumConsumptions(abilities []Ability) *uint64 {
	var total uint64
	for _, a := range abilities {
		consumptions := a.GetConsumptions()
		if consumptions == nil {
			// Unlimited ability => return nil to signify "infinite"
			return nil
		}
		total += uint64(*consumptions)
	}
	return &total
}
