package pos_operations

func RangeSplits(start int, len int, algin int) (ranges [][2]int) {
	ranges = make([][2]int, 0, (len+algin-1)/algin)
	currentStart := start
	for currentStart < start+len {
		currentEnd := (currentStart / algin) * algin
		if currentEnd <= currentStart {
			currentEnd += algin
		}
		if currentEnd > start+len {
			currentEnd = start + len
		}
		ranges = append(ranges, [2]int{currentStart, currentEnd - currentStart})
		currentStart = currentEnd
	}
	return ranges
}
