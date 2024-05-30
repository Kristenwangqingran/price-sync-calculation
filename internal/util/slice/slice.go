package slice

func ContainsUint32(slice []uint32, target uint32) bool {
	for _, v := range slice {
		if target == v {
			return true
		}
	}

	return false
}
