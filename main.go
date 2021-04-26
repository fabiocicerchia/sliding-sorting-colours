package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// LETTERS
const BlueLetters = "\033[34mB\033[0m"
const GreenLetters = "\033[32mG\033[0m"
const OrangeLetters = "\033[33mO\033[0m"
const RedLetters = "\033[31mR\033[0m"

// CIRCLES
const BlueCircles = "\033[34m●\033[0m"
const GreenCircles = "\033[32m●\033[0m"
const OrangeCircles = "\033[33m●\033[0m"
const RedCircles = "\033[31m●\033[0m"

// SYMBOLS
const BlueSymbols = "\033[34m▲\033[0m"
const GreenSymbols = "\033[32m■\033[0m"
const OrangeSymbols = "\033[33m●\033[0m"
const RedSymbols = "\033[31m★\033[0m"

var Colours = []string{}

const MaxItemsPerColumn = 3
const NumberColumns = 4
const PlaceholderSpots = NumberColumns + (NumberColumns - 1)

var DrawingInterval time.Duration = 10
var TotalSteps = 0

// STACK -----------------------------------------------------------------------
type Stack []string

func (s Stack) GetPosition(position int) string {
	if position >= s.Size() {
		return " "
	}

	return s[position]
}
func (s *Stack) MoveIn(item string) {
	if len(*s) == 0 {
		*s = make([]string, 0)
	}
	*s = append((*s), item)
}
func (s *Stack) MoveOut() string {
	n := len(*s) - 1
	item := s.GetPosition(n)
	*s = (*s)[:n]

	return item
}
func (s Stack) Size() int {
	return len(s)
}

// PLACEHOLDER -----------------------------------------------------------------
type Placeholder struct {
	Stack
}

func (p Placeholder) CouldPick(item string) bool {
	return p.GetPickPosition(item) > -1
}
func (p Placeholder) GetPickPosition(item string) int {
	// TODO: Check order on slot in placeholders and column

	for i := 0; i < PlaceholderSpots; i++ {
		if p.Stack.GetPosition(i) == item {
			return i
		}
	}

	return -1
}
func (p *Placeholder) Pick(item string) string {
	if i := p.GetPickPosition(item); i > -1 {
		p.Stack = append(p.Stack[:i], p.Stack[i+1:]...)
		return item
	}

	return ""
}

// COLUMN ----------------------------------------------------------------------
type Column struct {
	Stack
}

func (c Column) GetPosition(position int) string {
	return c.Stack.GetPosition(position)
}
func (c Column) HasFreeSlots() bool {
	return c.Stack.Size() < MaxItemsPerColumn
}
func (c Column) HasSameValueForPositionInColumn(position int, column Column) bool {
	return c.GetPosition(position) == column.GetPosition(position)
}
func (c Column) IsEmpty() bool {
	return c.Size() == 0
}
func (c Column) IsEqual(c2 Column) bool {
	if c.Stack.Size() != c2.Stack.Size() {
		return false
	}

	for item := 0; item < c2.Stack.Size(); item++ {
		if !c.HasSameValueForPositionInColumn(item, c2) {
			return false
		}
	}

	return true
}
func (c Column) IsSlotEmpty(position int) bool {
	return c.Stack.GetPosition(position) == " "
}
func (c *Column) MoveIn(item string) {
	c.Stack.MoveIn(item)
}
func (c *Column) MoveOut() string {
	return c.Stack.MoveOut()
}
func (c Column) Size() int {
	return c.Stack.Size()
}

// BOARD -----------------------------------------------------------------------
var CurrentBoard Board
var DesiredBoard Board

type Board struct {
	Placeholders  Placeholder
	Columns       []Column
	WorkingColumn int
}

func ReadItem() string {
	fmt.Print("Enter Item Colour: ")
	text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimRight(strings.ToUpper(text), "\n")
}

func (b *Board) addPlaceholder(item string) {
	b.Placeholders.MoveIn(item)
}
func (b *Board) Clone(b2 Board) {
	b.Init("none")

	b.Placeholders = b2.Placeholders

	for column := 0; column < len(b2.Columns); column++ {
		for spots := 0; spots < b2.Columns[column].Size(); spots++ {
			b.moveIntoColumn(column, b2.GetPosition(column, spots))
		}
	}
}
func (b Board) Draw(clean bool) {
	if clean {
		fmt.Printf("\033[1A")
		for i := 0; i < MaxItemsPerColumn+1; i++ {
			fmt.Printf("\033[1A")
		}
		time.Sleep(DrawingInterval * time.Millisecond)

		TotalSteps++
		fmt.Printf("--- STEP #%d\n", TotalSteps)
	}

	for i := 0; i < PlaceholderSpots; i++ {
		item := string(b.Placeholders.GetPosition(i + 1))

		if val := b.Placeholders.GetPosition(0); i == 0 && val != "" { // First Edge
			item = string(val)
		} else if val := b.Placeholders.GetPosition(1); i == PlaceholderSpots-1 && val != "" { // Second Edge
			item = string(val)
		}

		fmt.Printf("[%s] ", item)
	}

	fmt.Print("\n")
	for i := MaxItemsPerColumn - 1; i >= 0; i-- {
		fmt.Printf("   ") // offsetting column
		for column := 0; column < NumberColumns; column++ {
			fmt.Printf("[%s]   ", b.GetPosition(column, i))
		}
		fmt.Print("\n")
	}
}
func (b Board) GetPosition(column int, position int) string {
	return b.Columns[column].GetPosition(position)
}
func (b *Board) Init(mode string) {
	b.Columns = make([]Column, NumberColumns)

	if mode == "shuffle" {
		b.Shuffle()
	} else if mode != "none" {
		b.ReadPositions()
	}
}
func (b Board) IsColumnEmpty(column int) bool {
	return b.Columns[column].IsEmpty()
}
func (b Board) IsSlotEmpty(column int, position int) bool {
	return b.Columns[column].IsSlotEmpty(position)
}
func (b Board) LookupItem(item string) (int, int) {
	for column := b.WorkingColumn + 1; column < NumberColumns; column++ {
		for spots := MaxItemsPerColumn - 1; spots >= 0; spots-- {
			if b.Columns[column].Size() > spots && b.GetPosition(column, spots) == item {
				return column, spots
			}
		}
	}

	return -1, -1
}
func (b Board) LookupPlaceholderItem(item string) int {
	for spots := PlaceholderSpots - 1; spots >= 0; spots-- {
		if b.Placeholders.Size() > spots && b.Placeholders.GetPosition(spots) == item {
			return spots
		}
	}

	return -1
}
func (b *Board) MoveFromColumnIntoPlaceholder(column int) {
	if b.Placeholders.Size() > PlaceholderSpots {
		panic("Placeholder Overflow: too many placeholders")
	}

	b.addPlaceholder(b.Columns[column].MoveOut())
	b.Draw(true)
}
func (b *Board) MoveFromColumnToColumn(columnA int, columnB int) {
	b.moveIntoColumn(columnB, b.Columns[columnA].MoveOut())
	b.Draw(true)
}
func (b *Board) moveIntoColumn(column int, item string) {
	b.Columns[column].MoveIn(item)
}
func (b *Board) ReadPositions() {
	for column := 0; column < NumberColumns; column++ {
		fmt.Printf("Insert Each Items in Column #%d (from bottom to top)\n", column+1)
		for spots := 0; spots < MaxItemsPerColumn; spots++ {
			b.moveIntoColumn(column, ReadItem())
		}
	}
}
func (b *Board) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	var picks []string
	for column := 0; column < NumberColumns; column++ {
		for items := 0; items < MaxItemsPerColumn; items++ {
			picks = append(picks, string(Colours[column]))
		}
	}

	for column := 0; column < NumberColumns; column++ {
		for spots := 0; spots < MaxItemsPerColumn; spots++ {
			pos := rand.Intn(len(picks))
			item := picks[pos]
			picks = append(picks[:pos], picks[pos+1:]...)
			b.moveIntoColumn(column, item)
		}
	}
}

// -----------------------------------------------------------------------------
func flowColumn(nextUnalignedItem int, clonedBoard *Board, desiredColumn Column, column int) {
	for i := 1; i <= MaxItemsPerColumn-nextUnalignedItem; i++ {
		if !clonedBoard.IsColumnEmpty(column) && !clonedBoard.Columns[column].HasSameValueForPositionInColumn(nextUnalignedItem, desiredColumn) {
			clonedBoard.MoveFromColumnIntoPlaceholder(column)
		}
	}

	for i := 0; i < MaxItemsPerColumn; i++ {
		if clonedBoard.Columns[column].HasSameValueForPositionInColumn(i, desiredColumn) {
			continue
		}

		// check in placeholders
		pos := clonedBoard.LookupPlaceholderItem(desiredColumn.GetPosition(i))
		if pos != -1 {
			clonedBoard.moveIntoColumn(column, clonedBoard.Placeholders.Pick(desiredColumn.GetPosition(i)))
			clonedBoard.Draw(true)
			continue
		}

		// check in columns
		col, pos := clonedBoard.LookupItem(desiredColumn.GetPosition(i))
		if pos > -1 && pos < 2 {
			for i := 2; i > 0; i-- {
				if !clonedBoard.IsSlotEmpty(col, i) {
					clonedBoard.MoveFromColumnIntoPlaceholder(col)
				}
			}
		}
		if pos > -1 && pos <= 2 {
			clonedBoard.MoveFromColumnToColumn(col, column)
		}
	}
}
func flowPlaceholder(nextUnalignedItem int, clonedBoard *Board, desiredColumn Column, column int) {
	for i := nextUnalignedItem; i < MaxItemsPerColumn; i++ {
		if clonedBoard.Placeholders.CouldPick(desiredColumn.GetPosition(i)) {
			clonedBoard.moveIntoColumn(column, clonedBoard.Placeholders.Pick(desiredColumn.GetPosition(i)))
			clonedBoard.Draw(true)
			break
		}

		if col, pos := clonedBoard.LookupItem(desiredColumn.GetPosition(i)); pos != -1 {
			clonedBoard.MoveFromColumnToColumn(col, column)
			break
		}
	}
}
func getNextUnalignedItem(currentColumn Column, desiredColumn Column) int {
	nextUnalignedItem := 0
	for i := 0; i < MaxItemsPerColumn; i++ {
		if !currentColumn.HasSameValueForPositionInColumn(i, desiredColumn) {
			break
		}

		nextUnalignedItem++
	}

	return nextUnalignedItem
}
func initColourPalette(typeFlag string) []string {
	if typeFlag == "letters" {
		return []string{BlueLetters, GreenLetters, OrangeLetters, RedLetters}
	} else if typeFlag == "symbols" {
		return []string{BlueSymbols, GreenSymbols, OrangeSymbols, RedSymbols}
	}

	return []string{BlueCircles, GreenCircles, OrangeCircles, RedCircles}
}
func Solve(current *Board, desired Board) {
	wasNotSorted := false
	for column := 0; column < NumberColumns; column++ {
		if !current.Columns[column].IsEqual(desired.Columns[column]) {
			SolveColumn(current, desired, column)
			wasNotSorted = true
		}
	}

	if wasNotSorted {
		Solve(current, desired)
	}
}
func SolveColumn(current *Board, desired Board, column int) {
	desiredColumn := desired.Columns[column]

	nextUnalignedItem := getNextUnalignedItem(current.Columns[column], desiredColumn)
	// column already sorted
	if nextUnalignedItem == MaxItemsPerColumn {
		return
	}

	clonedBoard := &Board{
		WorkingColumn: column,
	}
	clonedBoard.Clone(*current)

	pickFromPlaceholders := nextUnalignedItem > 0 || clonedBoard.IsColumnEmpty(column)
	if pickFromPlaceholders && clonedBoard.Columns[column].HasFreeSlots() && clonedBoard.Placeholders.CouldPick(desiredColumn.GetPosition(nextUnalignedItem)) {
		flowPlaceholder(nextUnalignedItem, clonedBoard, desiredColumn, column)
	} else {
		flowColumn(nextUnalignedItem, clonedBoard, desiredColumn, column)
	}

	current.Clone(*clonedBoard)
}

// -----------------------------------------------------------------------------

func main() {
	var typeFlag = flag.String("type", "circles", "Type of drawing (letters, circles, symbols). Default: circles")
	var modeFlag = flag.String("mode", "shuffle", "Gaming mode (shuffle or input). Default: shuffle")
	var timeFlag = flag.Int("time", 10, "Drawing Speed. Default: 10ms")

	flag.Parse()
	DrawingInterval = time.Duration(*timeFlag)

	Colours = initColourPalette(*typeFlag)

	fmt.Println("SLIDING SORTING COLOURS")

	fmt.Println("---")
	fmt.Println("Starting Board")
	CurrentBoard.Init(*modeFlag)
	CurrentBoard.Draw(false)

	fmt.Println("---")
	fmt.Println("Desired Board")
	DesiredBoard.Init(*modeFlag)
	DesiredBoard.Draw(false)

	fmt.Println("Calculating...")
	CurrentBoard.Draw(false)
	Solve(&CurrentBoard, DesiredBoard)
}
