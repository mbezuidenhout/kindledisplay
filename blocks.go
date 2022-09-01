package main

type Block struct {
	X, Y          int
	Width, Height int
}

type Blocks struct {
	Blocks        [4]Block
	Width, Height int
}

/*
Returns a block layout based on the nrOfBlocks and PageOrientation
*/
func NewBlockLayout(nrOfBlocks int, PageOrientation Orientation) *Blocks {
	var OneBlockLayout Blocks = Blocks{}
	var TwoBlockLayout Blocks = Blocks{}
	var ThreeBlockLayout Blocks = Blocks{}
	var FourBlockLayout Blocks = Blocks{}

	OneBlockLayout.Blocks[0] = Block{X: 0, Y: 0, Width: 600, Height: 800}
	TwoBlockLayout.Blocks[0] = Block{X: 0, Y: 0, Width: 600, Height: 400}
	TwoBlockLayout.Blocks[1] = Block{X: 0, Y: 400, Width: 600, Height: 400}
	ThreeBlockLayout.Blocks[0] = Block{X: 0, Y: 0, Width: 600, Height: 400}
	ThreeBlockLayout.Blocks[1] = Block{X: 0, Y: 400, Width: 300, Height: 400}
	ThreeBlockLayout.Blocks[2] = Block{X: 300, Y: 400, Width: 300, Height: 400}
	FourBlockLayout.Blocks[0] = Block{X: 0, Y: 0, Width: 300, Height: 400}
	FourBlockLayout.Blocks[1] = Block{X: 300, Y: 0, Width: 300, Height: 400}
	FourBlockLayout.Blocks[2] = Block{X: 0, Y: 400, Width: 300, Height: 400}
	FourBlockLayout.Blocks[3] = Block{X: 300, Y: 400, Width: 300, Height: 400}

	var b *Blocks
	switch nrOfBlocks {
	case 1:
		b = &OneBlockLayout
	case 2:
		b = &TwoBlockLayout
	case 3:
		b = &ThreeBlockLayout
	case 4:
		b = &FourBlockLayout
	}

	b.Width = 600
	b.Height = 800

	// Transpose Y, X, Width and Height
	if PageOrientation == Landscape {
		b.Width = 800
		b.Height = 600
		for k, _ := range b.Blocks {
			height := b.Blocks[k].Height
			b.Blocks[k].Height = b.Blocks[k].Width
			b.Blocks[k].Width = height
			y := b.Blocks[k].Y
			b.Blocks[k].Y = b.Blocks[k].X
			b.Blocks[k].X = y
		}
	}

	return b
}
