package entities

func SumConsumptions(abilities []Ability) *uint64 {
	var total uint64
	for _, a := range abilities {
		consumptions := a.GetConsumptions()
		if consumptions == nil {
			// Une capacité infinie => on retourne nil pour signifier "infini"
			return nil
		}
		total += uint64(*consumptions)
	}
	return &total
}
