// internal/uicommon/menu.go
package uicommon

// ComputeButtonPositions вычисляет X, Y для списка кнопок, центрируя блок по вертикали
func ComputeButtonPositions(items []string, buttonWidth, buttonHeight, gap float32, screenW, screenH int, minY float32) (buttonX float32, buttonY []float32) {
	buttonX = (float32(screenW) - buttonWidth) / 2
	totalH := float32(len(items))*buttonHeight + float32(len(items)-1)*gap
	startY := (float32(screenH) - totalH) / 2
	if startY < minY {
		startY = minY
	}
	buttonY = make([]float32, len(items))
	for i := range items {
		buttonY[i] = startY + float32(i)*(buttonHeight+gap)
	}
	return buttonX, buttonY
}
