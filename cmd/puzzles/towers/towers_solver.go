package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

type RawBoard [][]int

type RawGameInput struct {
	Constraints Constraints

	Board RawBoard
}

type SideConstraints []int

type Constraints struct {
	Left   SideConstraints
	Right  SideConstraints
	Top    SideConstraints
	Bottom SideConstraints
}

// Cell represents a collection of information relevant
// to a single cell in the board.
type Cell struct {

	// SolvedValue represents the solved value of the cell.
	// A value of 0 indicates it is unsolved.
	SolvedValue int

	// PossibleValues represents the possible values that the
	// cell can take. It is a mutable structure that evolves
	// through the solution, and eventually reduces to just
	// a single value, in which case, the cell can be considered
	// as solved.
	PossibleValues map[int]bool
}

func (c Cell) IsSolved() bool {
	return c.SolvedValue > 0
}

func (c Cell) String() string {
	if c.IsSolved() {
		return fmt.Sprintf("%d", c.SolvedValue)
	}

	possibleValues := c.GetPossibleValuesSorted()
	possibleValueStrs := make([]string, 0)
	for _, possibleValue := range possibleValues {
		possibleValueStrs = append(possibleValueStrs, fmt.Sprintf("%d", possibleValue))
	}

	return fmt.Sprintf("(%s)", strings.Join(possibleValueStrs, ","))
}

// IsCellUnsolvedButImmediatelySolvable returns if the cell can be immediately
// solved but hasn't yet been marked as solved. This happens if there is only one
// value in cell.PossibleValues but the cell.SolvedValue is unset (i.e. set to 0).
func (c Cell) IsCellUnsolvedButImmediatelySolvable() bool {
	return len(c.PossibleValues) == 1 && c.SolvedValue == 0
}

func (c Cell) GetOnlyPossibleValue() int {
	if len(c.PossibleValues) != 1 {
		panic(fmt.Sprintf("attempting to access only possible value from cell, but it has %d possible values", len(c.PossibleValues)))
	}
	for k := range c.PossibleValues {
		return k
	}
	panic("cannot happen")
}

func (c Cell) GetPossibleValuesSorted() []int {
	possibleValues := make([]int, 0, len(c.PossibleValues))
	for possibleValue := range c.PossibleValues {
		possibleValues = append(possibleValues, possibleValue)
	}
	slices.SortFunc(possibleValues, func(a, b int) int { return a - b })
	return possibleValues
}

func (c Cell) GetPossibleValuesMinMax() MinMax {

	minVal := 0
	maxVal := 0
	for possibleValue := range c.PossibleValues {
		if minVal == 0 || possibleValue < minVal {
			minVal = possibleValue
		}
		if maxVal == 0 || possibleValue > maxVal {
			maxVal = possibleValue
		}
	}
	return MinMax{
		Min: minVal,
		Max: maxVal,
	}
}

type MinMax struct {
	Min int
	Max int
}

// Combine returns a min-max value that contains
// the minimum of the two minimums, and
// the maximum of the two maximums.
//
// Ignores 0 values, since it treats them as "missing info".
func (m MinMax) Combine(o MinMax) MinMax {

	minVal := m.Min
	maxVal := m.Max

	if o.Min != 0 && o.Min < m.Min {
		minVal = o.Min
	}

	if o.Max != 0 && o.Max > m.Max {
		maxVal = o.Max
	}

	return MinMax{
		Min: minVal,
		Max: maxVal,
	}
}

// Board represents the entire game board, along with the side constraints
// that define the problem.
type Board struct {
	Cells       [][]*Cell
	Constraints Constraints
}

func NewBoard(cells [][]*Cell, constraints Constraints) Board {
	return Board{
		Cells:       cells,
		Constraints: constraints,
	}
}

// Size returns the size of any side of the board.
// All sides are equal in size.
func (b Board) Size() int {
	return len(b.Cells)
}

func (b Board) String() string {

	N := b.Size()

	// First, construct the strings that represents the cells in each row.
	// Keep track of the widest string representation found across the board.
	boardStrings := make([][]string, 0, N+2)
	maxStrLength := 0

	// row = -1 represents the top constraints row.
	// row = N represents the bottom constraints row.
	for row := -1; row <= N; row++ {
		rowStrings := make([]string, 0, N+2)

		// col = -1 represents the left constraints column.
		// col = N represents the right constraints column.
		for col := -1; col <= N; col++ {

			str := ""
			if row == -1 {
				if col >= 0 && col < N {
					str = fmt.Sprintf("%d", b.Constraints.Top[col])
				}
			} else if row == N {
				if col >= 0 && col < N {
					str = fmt.Sprintf("%d", b.Constraints.Bottom[col])
				}
			} else {
				if col == -1 {
					str = fmt.Sprintf("%d", b.Constraints.Left[row])
				} else if col == N {
					str = fmt.Sprintf("%d", b.Constraints.Right[row])
				} else {
					str = b.Cells[row][col].String()
				}

			}

			rowStrings = append(rowStrings, str)
			if len(str) > maxStrLength {
				maxStrLength = len(str)
			}
		}

		boardStrings = append(boardStrings, rowStrings)
	}
	maxStrLength += 2 // give padding even to the longest string.

	// Rewrite the strings so that they all have the same length.
	// And merge each row into a single string.
	mergedRowStrings := make([]string, 0, N+2)
	for row := -1; row <= N; row++ {
		for col := -1; col <= N; col++ {
			str := boardStrings[row+1][col+1]

			lenPadding := maxStrLength - len(str)

			paddedStr := fmt.Sprintf(
				"%s%s%s",
				strings.Repeat(" ", lenPadding/2), str, strings.Repeat(" ", lenPadding-lenPadding/2))
			boardStrings[row+1][col+1] = paddedStr
		}

		// We don't want the demarcator to show for the top and bottom constraint rows.
		demarcator := "|"
		if row == -1 || row == N {
			demarcator = " "
		}

		mergedRowStrings = append(mergedRowStrings, strings.Join(boardStrings[row+1], demarcator))
	}

	// Now, merge the merged row strings into a single string.
	// To do this, we need the row demarcator. For most part, this will be the '-' character.
	// However, we don't want the demarcator to show for the left and right constraint columns.
	paddedStrLen := len(boardStrings[0][0])
	emptyStrPaddedLen := strings.Repeat(" ", paddedStrLen)
	rowDemarcator := fmt.Sprintf(
		"\n%s%s%s\n",
		emptyStrPaddedLen,
		strings.Repeat("-", N+1 /* to account for the column demarcators */ +N*paddedStrLen /* for the row contents */),
		emptyStrPaddedLen)

	return strings.Join(mergedRowStrings, rowDemarcator)
}

type GameSolution struct {
	FinalBoard Board
	Steps      []string
}

type Solver struct {
	board          Board
	numSolvedCells int

	solutionSteps []string
}

func NewSolver(rgi RawGameInput) *Solver {

	N := len(rgi.Board)

	cells := make([][]*Cell, 0, N)
	for row := 0; row < N; row++ {
		rowCells := make([]*Cell, 0, N)
		for col := 0; col < N; col++ {

			cellPossibleValues := make(map[int]bool, 0)
			for val := 1; val <= N; val++ {
				cellPossibleValues[val] = true
			}

			// We are not encoding the known values yet. We'll do that
			// below separately.
			rowCells = append(rowCells, &Cell{
				PossibleValues: cellPossibleValues,
			})
		}

		cells = append(cells, rowCells)
	}

	b := NewBoard(cells, rgi.Constraints)

	s := &Solver{
		board:         b,
		solutionSteps: make([]string, 0),
	}

	// Now, we'll insert specific values at cells based on the raw input.
	for row := 0; row < N; row++ {
		for col := 0; col < N; col++ {
			val := rgi.Board[row][col]
			if val > 0 {
				s.setSolvedValue(row, col, val)
			}
		}
	}

	return s
}

// IsTerminal indicates if the solver has solved the board it has been given.
func (s Solver) Solved() bool {
	N := s.board.Size()
	return s.numSolvedCells == N*N
}

func (s *Solver) Solve() GameSolution {

	numSolutionStepsAtStartOfIteration := len(s.solutionSteps)

	// This is a special strategy that we only need to set once.
	s.setMaxNumbersAlongEdge()

	// On each iteration, try various strategies to narrow down the possibilities.
	// We'll use a bunch of ad-hoc rules.
	for !s.Solved() {

		s.eliminatePossibleValuesBasedOnDistanceFromEdge()

		s.eliminatePossibleValuesFurtherAwayFromEdgeWhenSideConstraintIsEqualTo2()

		s.eliminatePossibleValuesThatViolateSideConstraint()

		s.solveHiddenSingles()

		if len(s.solutionSteps) == numSolutionStepsAtStartOfIteration {
			// At this point, we can check to see if the board isn't solved, and just
			// backtrack. But skipping that because I'm lazy.

			break
		}
		numSolutionStepsAtStartOfIteration = len(s.solutionSteps)
	}

	return GameSolution{
		FinalBoard: s.board,
		Steps:      s.solutionSteps,
	}
}

// setSolvedValue sets the value at a given cell, and
// erases the possibilities from other cells in the same
// row and column. If that causes those cells to become immediately
// solveable, recursively repeats the process.
func (s *Solver) setSolvedValue(row, col, val int) {

	cell := s.board.Cells[row][col]

	if cell.SolvedValue == val {
		// Noop. We have already done this in the past.
		return
	}

	// Fix the cell.
	s.numSolvedCells++
	cell.SolvedValue = val

	possibleValuesToRemove := make([]int, 0)
	for possibleValue := range cell.PossibleValues {
		if possibleValue != val {
			possibleValuesToRemove = append(possibleValuesToRemove, possibleValue)
		}
	}
	for _, valueToRemove := range possibleValuesToRemove {
		delete(cell.PossibleValues, valueToRemove)
	}

	N := s.board.Size()

	valsToRemove := []int{val}

	// Fix the row.
	s.registerSolutionStep(solutionStep_RemovingPossibilitiesFromRow(row, valsToRemove))
	for c := 0; c < N; c++ {
		if c == col {
			continue // this is the solved cell.
		}
		s.removePossibleValuesFromCell(row, c, valsToRemove)
	}

	// Fix the column.
	s.registerSolutionStep(solutionStep_RemovingPossibilitiesFromColumn(col, valsToRemove))
	for r := 0; r < N; r++ {
		if r == row {
			continue // this is the solved cell.
		}
		s.removePossibleValuesFromCell(r, col, valsToRemove)
	}
}

func (s *Solver) removePossibleValuesFromCell(row, col int, valuesToRemove []int) {

	cell := s.board.Cells[row][col]

	// Remove possible values from cell.
	for _, valueToRemove := range valuesToRemove {
		delete(cell.PossibleValues, valueToRemove)
	}

	// Recursive step: And if that make the cell a naked single, apply the change recursively.
	if cell.IsCellUnsolvedButImmediatelySolvable() {
		onlyPossibleValue := cell.GetOnlyPossibleValue()
		s.registerSolutionStep(solutionStep_NakedSingle(row, col, onlyPossibleValue))
		s.setSolvedValue(row, col, onlyPossibleValue)
	}
}

func solutionStep_RemovingPossibilitiesFromRow(row int, vals []int) string {
	return fmt.Sprintf("Removing possible values %s from (row %d) because these values are already present in row", strings.Join(strings.Fields(fmt.Sprint(vals)), ","), row)
}

func solutionStep_RemovingPossibilitiesFromColumn(col int, vals []int) string {
	return fmt.Sprintf("Removing possible values %s from (column %d) because these values are already present in column", strings.Join(strings.Fields(fmt.Sprint(vals)), ","), col)
}

func solutionStep_NakedSingle(row, col, val int) string {
	return fmt.Sprintf("Marking naked single in (row %d, column %d) with val %d as solved", row, col, val)
}

func (s *Solver) setMaxNumbersAlongEdge() {

	N := s.board.Size()

	constraintGetters := []func() SideConstraints{
		func() SideConstraints { return s.board.Constraints.Left },
		func() SideConstraints { return s.board.Constraints.Right },
		func() SideConstraints { return s.board.Constraints.Top },
		func() SideConstraints { return s.board.Constraints.Bottom },
	}

	rowColGetters := []func(idx int) (int, int){
		func(idx int) (int, int) { return idx, 0 },
		func(idx int) (int, int) { return idx, N - 1 },
		func(idx int) (int, int) { return 0, idx },
		func(idx int) (int, int) { return N - 1, idx },
	}

	for dir := 0; dir < 4; dir++ {
		constraints := constraintGetters[dir]()
		rowColGetter := rowColGetters[dir]
		for idx, constraint := range constraints {
			if constraint == 1 {
				row, col := rowColGetter(idx)
				s.registerSolutionStep(solutionStep_MaxValueAlongEdge(dir, row, col, N))
				s.setSolvedValue(row, col, N)
			}
		}
	}
}

func solutionStep_MaxValueAlongEdge(dir, row, col, val int) string {
	return fmt.Sprintf("Setting value %d at (row %d column %d) due to side %s constraint of 1", val, row, col, dirLabel(dir))
}

func (s *Solver) eliminatePossibleValuesBasedOnDistanceFromEdge() {
	N := s.board.Size()

	rowColGetters := []func(idx, distance int) (int, int){
		func(idx, distance int) (int, int) { return idx, distance },
		func(idx, distance int) (int, int) { return idx, N - 1 - distance },
		func(idx, distance int) (int, int) { return distance, idx },
		func(idx, distance int) (int, int) { return N - 1 - distance, idx },
	}

	constraintGetters := []func(idx int) int{
		func(idx int) int { return s.board.Constraints.Left[idx] },
		func(idx int) int { return s.board.Constraints.Right[idx] },
		func(idx int) int { return s.board.Constraints.Top[idx] },
		func(idx int) int { return s.board.Constraints.Bottom[idx] },
	}

	for dir := 0; dir < 4; dir++ {
		rowColGetter := rowColGetters[dir]
		constraintGetter := constraintGetters[dir]

		for idx := 0; idx < N; idx++ {
			constraint := constraintGetter(idx)
			for distance := 0; distance < N; distance++ {
				row, col := rowColGetter(idx, distance)
				cell := s.board.Cells[row][col]
				if cell.IsSolved() {
					if cell.SolvedValue == N {
						break
					}

					continue
				}

				maxValPossible := N - constraint + 1 + distance
				if maxValPossible >= N {
					continue
				}

				valuesToRemove := make([]int, 0)
				for val := range cell.PossibleValues {
					if val > maxValPossible {
						valuesToRemove = append(valuesToRemove, val)
					}
				}

				if len(valuesToRemove) == 0 {
					continue
				}

				s.registerSolutionStep(
					solutionStep_RemovingPossibleValuesDueToDistanceFromEdge(row, col, dir, constraint, distance, valuesToRemove))
				s.removePossibleValuesFromCell(row, col, valuesToRemove)

			}
		}
	}
}

func solutionStep_RemovingPossibleValuesDueToDistanceFromEdge(row, col, dir, constraint, distance int, vals []int) string {
	return fmt.Sprintf(
		"Removing possible values %s from (row %d, column %d) because "+
			"(a) the cell is at a distance %d from the edge and "+
			"(b) %s side constraint value of %d",
		strings.Join(strings.Fields(fmt.Sprint(vals)), ","), row, col, distance, dirLabel(dir), constraint)
}

func (s *Solver) eliminatePossibleValuesFurtherAwayFromEdgeWhenSideConstraintIsEqualTo2() {
	N := s.board.Size()

	rowColGetters := []func(idx, distance int) (int, int){
		func(idx, distance int) (int, int) { return idx, distance },
		func(idx, distance int) (int, int) { return idx, N - 1 - distance },
		func(idx, distance int) (int, int) { return distance, idx },
		func(idx, distance int) (int, int) { return N - 1 - distance, idx },
	}

	constraintGetters := []func(idx int) int{
		func(idx int) int { return s.board.Constraints.Left[idx] },
		func(idx int) int { return s.board.Constraints.Right[idx] },
		func(idx int) int { return s.board.Constraints.Top[idx] },
		func(idx int) int { return s.board.Constraints.Bottom[idx] },
	}

	for dir := 0; dir < 4; dir++ {
		rowColGetter := rowColGetters[dir]
		constraintGetter := constraintGetters[dir]

		for idx := 0; idx < N; idx++ {
			constraint := constraintGetter(idx)
			if constraint != 2 {
				continue
			}

			minDistForN, maxDistForN := s.getDistanceRangeForNumber(rowColGetter, idx, N)

			largestNumberThatComesBeforeN := 0 // not found.
			minDistForK := 0                   // not found

			// We want to find the largest number that comes before N as viewed from this
			// direction / side. And we want to eliminate it from those cells which are
			// not at a distance 0 from the edge.
			for K := N - 1; K >= 1; K-- {
				minDistForK, _ = s.getDistanceRangeForNumber(rowColGetter, idx, K)
				if minDistForK >= maxDistForN {
					continue
				}

				largestNumberThatComesBeforeN = K
				break
			}

			if largestNumberThatComesBeforeN == 0 {
				continue
			}

			if minDistForK < 0 {
				// Weird. There is a number cannot be present in some row or column. The puzzle is unsolvable.
				return
			}

			for distance := minDistForK; distance <= minDistForN; distance++ {
				if distance == 0 {
					continue
				}

				row, col := rowColGetter(idx, distance)
				cell := s.board.Cells[row][col]
				if cell.IsSolved() {
					continue
				}

				if _, ok := cell.PossibleValues[largestNumberThatComesBeforeN]; ok {
					s.registerSolutionStep(
						solutionStep_RemovingPossibleValuesFurtherAwayFromEdgeDueToSideConstraintOf2(row, col, dir, largestNumberThatComesBeforeN))
					s.removePossibleValuesFromCell(row, col, []int{largestNumberThatComesBeforeN})
				}
			}
		}
	}
}

// getDistanceRangeForNumber gets the range of distances from a side within which a given target value
// is bound for a given index (idx) on that side.
func (s Solver) getDistanceRangeForNumber(rowColGetter func(idx, distance int) (int, int), idx, targetVal int) (int, int) {

	N := s.board.Size()

	minDist := -1
	maxDist := -1

	for distance := 0; distance < N; distance++ {
		row, col := rowColGetter(idx, distance)
		cell := s.board.Cells[row][col]

		if cell.IsSolved() {
			if cell.SolvedValue == targetVal {
				return distance, distance
			}

			continue
		}

		for val := range cell.PossibleValues {
			if val != targetVal {
				continue
			}

			if minDist == -1 {
				minDist = distance
			}
			maxDist = distance
		}
	}

	return minDist, maxDist
}

func solutionStep_RemovingPossibleValuesFurtherAwayFromEdgeDueToSideConstraintOf2(row, col, dir, val int) string {
	return fmt.Sprintf(
		"Removing possible value %d from (row %d column %d) because "+
			"of the %s side constraint value of 2",
		val, row, col, dirLabel(dir))
}

func (s *Solver) eliminatePossibleValuesThatViolateSideConstraint() {

	N := s.board.Size()

	rowColGetters := []func(idx, distance int) (int, int){
		func(idx, distance int) (int, int) { return idx, distance },
		func(idx, distance int) (int, int) { return idx, N - 1 - distance },
		func(idx, distance int) (int, int) { return distance, idx },
		func(idx, distance int) (int, int) { return N - 1 - distance, idx },
	}

	constraintGetters := []func(idx int) int{
		func(idx int) int { return s.board.Constraints.Left[idx] },
		func(idx int) int { return s.board.Constraints.Right[idx] },
		func(idx int) int { return s.board.Constraints.Top[idx] },
		func(idx int) int { return s.board.Constraints.Bottom[idx] },
	}

	for dir := 0; dir < 4; dir++ {
		rowColGetter := rowColGetters[dir]
		constraintGetter := constraintGetters[dir]

		for idx := 0; idx < N; idx++ {
			constraint := constraintGetter(idx)

			for distance := 0; distance < N; distance++ {
				row, col := rowColGetter(idx, distance)
				cell := s.board.Cells[row][col]
				if cell.IsSolved() {
					continue
				}

				for val := range cell.PossibleValues {
					// Assume we set this cell to this value.
					if s.doesAssumedValueViolateConstraint(rowColGetter, idx, constraint, distance, val) {
						s.registerSolutionStep(solutionStep_RemovingPossibleValueDueToSideConstraint(row, col, dir, constraint, val))
						s.removePossibleValuesFromCell(row, col, []int{val})
					}
				}

			}
		}
	}
}

func (s *Solver) doesAssumedValueViolateConstraint(rowColGetter func(idx, distance int) (int, int), idx, constraint, distanceOfSetCell, assumedValForSetCell int) bool {

	N := s.board.Size()

	// Create a copy of the possible values assuming that we set a specific value for the given cell.
	possibleValsByDistance := s.createCopyOfPossibleValuesWithAssumedValue(rowColGetter, idx, distanceOfSetCell, assumedValForSetCell)

	numVisibleForSure := 0
	maxPossibleValSoFar := 0
	maxPossibleValueBasedOnSureValue := false
	visibleStatusKnownForAll := true

	// fmt.Printf("Checking side constraint %d at idx %d with possible values %v\n", constraint, idx, possibleValsByDistance)

	for distance := 0; distance < N; distance++ {
		possibleVals := possibleValsByDistance[distance]

		var minVal, maxVal int
		for val := range possibleVals {
			if minVal == 0 || val < minVal {
				minVal = val
			}
			if val > maxVal {
				maxVal = val
			}
		}

		isSureValue := minVal == maxVal
		isVisibleForSure := minVal > maxPossibleValSoFar
		maybeVisible := maxVal > maxPossibleValSoFar

		isInvisibleForSure := maxVal < maxPossibleValSoFar && maxPossibleValueBasedOnSureValue

		if isVisibleForSure {
			// fmt.Printf("Val = %d, max val so far = %d\n", minVal, maxPossibleValSoFar)
			numVisibleForSure++
			maxPossibleValSoFar = minVal

			if isSureValue {
				maxPossibleValueBasedOnSureValue = true
			}
		}
		if maybeVisible {
			maxPossibleValSoFar = maxVal
		}

		if isSureValue && minVal == N {
			break // No point checking further.
		}

		if !isVisibleForSure && !isInvisibleForSure {
			visibleStatusKnownForAll = false
		}
	}

	if numVisibleForSure > constraint {
		// fmt.Printf("Num visible for sure: %d\n", numVisibleForSure)
		return true
	}

	if numVisibleForSure < constraint && visibleStatusKnownForAll {
		// fmt.Printf("Num visible for sure: %d, and all visibility status fully known\n", numVisibleForSure)
		return true
	}

	return false
}

func (s *Solver) createCopyOfPossibleValuesWithAssumedValue(rowColGetter func(idx, distance int) (int, int), idx, distanceOfSetCell, assumedValForSetCell int) []map[int]bool {

	N := s.board.Size()

	possibleValsByDistance := make([]map[int]bool, N)

	for distance := 0; distance < N; distance++ {
		row, col := rowColGetter(idx, distance)
		cell := s.board.Cells[row][col]

		possibleValsCpy := make(map[int]bool, 0)

		if distance == distanceOfSetCell {
			possibleValsCpy[assumedValForSetCell] = true
		} else {
			// Copy the original vals.
			for val := range cell.PossibleValues {
				possibleValsCpy[val] = true
			}
		}

		possibleValsByDistance[distance] = possibleValsCpy
	}

	assumedVals := []int{assumedValForSetCell}
	for len(assumedVals) > 0 {
		assumedVal := assumedVals[len(assumedVals)-1]
		assumedVals = assumedVals[:len(assumedVals)-1]

		for distance := 0; distance < N; distance++ {
			possibleVals := possibleValsByDistance[distance]

			if len(possibleVals) == 1 { // either already solved cell, or assumed val.
				continue
			}

			delete(possibleVals, assumedVal)

			if len(possibleVals) == 1 {
				for possibleVal := range possibleVals {
					assumedVals = append(assumedVals, possibleVal)
					break
				}
			}
		}
	}

	return possibleValsByDistance
}

func solutionStep_RemovingPossibleValueDueToSideConstraint(row, col, dir, constraint, val int) string {
	return fmt.Sprintf(
		"Removing possible value %d from (row %d, column %d) because "+
			"of the %s side constraint value of %d",
		val, row, col, dirLabel(dir), constraint)
}

// solveHiddenSingles attempts to see if there is some row or column
// for which some number appears only once in that row or column respectively.
// If so, it sets that as a solved value in the respective cell, and applies
// any changes recursively.
func (s *Solver) solveHiddenSingles() {

	N := s.board.Size()

	// First attempt along each row.
	for row := 0; row < N; row++ {
		countAlongRow := make(map[int][]int, 0)
		for col := 0; col < N; col++ {
			cell := s.board.Cells[row][col]
			if cell.IsSolved() {
				continue
			}
			for val := range cell.PossibleValues {
				countAlongRow[val] = append(countAlongRow[val], col)
			}
		}

		for val, cols := range countAlongRow {
			if len(cols) == 1 {
				// Found hidden single.
				col := cols[0]
				s.registerSolutionStep(solutionStep_HiddenSingleInRow(row, col, val))
				s.setSolvedValue(row, col, val)
			}
		}
	}

	// Next, attempt along each column.
	for col := 0; col < N; col++ {
		countAlongColumn := make(map[int][]int, 0)
		for row := 0; row < N; row++ {
			cell := s.board.Cells[row][col]
			if cell.IsSolved() {
				continue
			}
			for val := range cell.PossibleValues {
				countAlongColumn[val] = append(countAlongColumn[val], row)
			}
		}

		for val, rows := range countAlongColumn {
			if len(rows) == 1 {
				// Found hidden single.
				row := rows[0]
				s.registerSolutionStep(solutionStep_HiddenSingleInColumn(row, col, val))
				s.setSolvedValue(row, col, val)
			}
		}
	}
}

func solutionStep_HiddenSingleInRow(row, col, val int) string {
	return fmt.Sprintf("Marking hidden single in (row %d, column %d) with value %d as solved value", row, col, val)
}

func solutionStep_HiddenSingleInColumn(row, col, val int) string {
	return fmt.Sprintf("Marking hidden single in (column %d, row %d) with value %d as solved value", col, row, val)
}

func (s *Solver) registerSolutionStep(step string) {
	fmt.Println(step)
	fmt.Println("-------------")
	fmt.Println(s.board.String())
	fmt.Println()
	s.solutionSteps = append(s.solutionSteps, step)
}

func dirLabel(dir int) string {
	switch dir {
	case 0:
		return "left -> right"
	case 1:
		return "right -> left"
	case 2:
		return "top -> bottom"
	case 3:
		return "bottom -> top"
	default:
		panic("impossible direction")
	}
}

func main() {

	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("Unable to read from stdin: %s", err.Error())
		os.Exit(1)
	}

	var gi RawGameInput
	err = json.Unmarshal(bytes, &gi)
	if err != nil {
		fmt.Printf("Unable to unmarshal game input: %s", err.Error())
		os.Exit(1)
	}

	s := NewSolver(gi)
	soln := s.Solve()

	fmt.Println("Final board state is as follows...")
	fmt.Println("-------------")
	fmt.Println(soln.FinalBoard.String())
}
