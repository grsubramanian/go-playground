package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// GameConfig represents the basic configuration related to a ballsort puzzle game.
type GameConfig struct {
	NumContainers           int
	MaxNumBallsPerContainer int

	Colors []string
}

// Container is a set of balls stacked one on top of the other.
//
// The values encode the colors of the balls, which are represented as
// integers in the range [1, N], where N is the number of colors as per the
// game configuration.
//
// The ball (if any) in the 0th position is at the bottom of the container.
type Container []int

func (c Container) NumBalls() int {
	return len(c)
}

func (c Container) IsSameColor() bool {
	if c.NumBalls() == 0 {
		return true
	}

	ball1Color := c[0]

	for _, ballColor := range c {
		if ballColor != ball1Color {
			return false
		}
	}

	return true
}

// CanonicalValue gets the decimal equivalent of the k-digit base-(n+1) integer
// that is represented by the balls present in the container, where
// 'k' is the maximum number of balls per container, and 'n' is the number of
// colors.
//
// The idea of canonicalization is useful in the context of graph search, where
// we want to search through game states, but want to avoid redundant searches.
//
// We'll treat the lowermost ball in the container as the unit's place.
//
// An empty ball position gets treated as a 0. Therefore, an empty container gets
// treated as a 0.
func (c Container) CanonicalValue(cfg GameConfig) int {

	val := 0
	multiplier := 1
	base := len(cfg.Colors) + 1

	for _, ballColor := range c {
		val += (multiplier * ballColor)
		multiplier *= base
	}

	return val
}

// ContainerFromCanonicalValue is the inverse function of Container.CanonicalValue.
func ContainerFromCanonicalValue(val, maxNumBallsPerContainer, numColors int) Container {

	balls := make([]int, 0)

	base := numColors + 1

	for val > 0 {
		nextBallColor := val % base
		balls = append(balls, nextBallColor)
		val -= nextBallColor
		val /= base
	}

	return Container(balls)
}

// GameState represents the state of containers.
type GameState struct {
	Containers []Container

	// For performance reasons, we store a cached value of the canonical form
	// of the game state.
	cachedCanonicalForm *GameStateCanonicalForm
}

// IsTerminal indicates if the state represents a solved state.
// This happens when each of the non-empty containers contains
// balls of exactly one color.
func (s GameState) IsTerminal(cfg GameConfig) bool {

	for _, container := range s.Containers {
		// The container must contain balls of the same color.
		if !container.IsSameColor() {
			return false
		}

		// The container must either be empty or full.
		numBallsInContainer := container.NumBalls()
		if numBallsInContainer > 0 && numBallsInContainer < cfg.MaxNumBallsPerContainer {
			return false
		}
	}

	return true
}

// GameStateCanonicalForm represents a canonical form of the game state.
// We can treat two game states that have the same set of containers but
// in a different order as identical. Further, for each container, we
// can reduce it to an integer as defined by the Container.CanonicalValue()
// function. Finally, we can treat the array of canonical values as a string.
type GameStateCanonicalForm string

// CanonicalForm gets the canonical form of the game state.
//
// For performance reasons, caches the canonical form.
func (s *GameState) CanonicalForm(cfg GameConfig) GameStateCanonicalForm {

	if s.cachedCanonicalForm != nil {
		return *s.cachedCanonicalForm
	}

	canonicalValues := make([]int, 0)

	for _, container := range s.Containers {
		canonicalValue := container.CanonicalValue(cfg)
		canonicalValues = append(canonicalValues, canonicalValue)
	}

	slices.SortFunc(canonicalValues, func(i, j int) int { return i - j })

	canonicalValueStrs := make([]string, len(canonicalValues))
	for i := 0; i < len(canonicalValues); i++ {
		canonicalValueStrs[i] = fmt.Sprintf("%d", canonicalValues[i])
	}

	canonicalForm := GameStateCanonicalForm(strings.Join(canonicalValueStrs, ","))
	s.cachedCanonicalForm = &canonicalForm
	return canonicalForm
}

// CloneWithBallsMoved creates a similar game state to the given one, except a certain number of balls
// have been moved from a certain container to another.
//
// It is the caller's responsibility to pass legitimate values.
func (s GameState) CloneWithBallsMoved(fromContainerIdx, toContainerIdx int, numBallsToMove int) GameState {

	numContainers := len(s.Containers)

	clonedContainers := make([]Container, 0, numContainers)

	for i := 0; i < numContainers; i++ {

		container := s.Containers[i]

		clonedContainer := make([]int, len(container))
		copy(clonedContainer, container)

		clonedContainers = append(clonedContainers, clonedContainer)
	}

	fromContainer := clonedContainers[fromContainerIdx]
	numBallsInFromContainer := fromContainer.NumBalls()

	clonedContainers[toContainerIdx] = append(
		clonedContainers[toContainerIdx],
		clonedContainers[fromContainerIdx][numBallsInFromContainer-numBallsToMove:numBallsInFromContainer]...)
	clonedContainers[fromContainerIdx] = fromContainer[:numBallsInFromContainer-numBallsToMove]

	return GameState{
		Containers: clonedContainers,
	}
}

type GameSolution[T any] struct {
	// Transitions represents the set of game state transitions required to go from
	// the initial game state to the final game state.
	//
	// If nil, represents, an unsolved / unsolveable game.
	Transitions []GameStateTransition

	// Stats contain some statistics of the game solution process.
	// The type of object depends on the type of game solver.
	Stats T
}

type GameStateTransition struct {
	FromContainerIdx int
	ToContainerIdx   int
	NumBalls         int

	FromGameState GameState
	ToGameState   GameState
}

// GameSolver represents an arbitrary game solver than can solve the ballsort game.
type GameSolver[T any] interface {
	Solve(cfg GameConfig, startingGameState GameState) GameSolution[T]
}

type DFSSearchStats struct {
	NumVisitedGameStates  int
	NumExploredGameStates int
}

type DFSGameSolver struct {

	// 0] a map providing the game state that was first responsible
	// for a given canonical form, during the DFS exploration.
	gameStateForGivenCanonicalForm map[GameStateCanonicalForm]GameState

	// 1] a map indicating the game state transition that was most recently
	// responsible for visiting a game state.
	gameStateTransitionForGivenCanonicalForm map[GameStateCanonicalForm]GameStateTransition

	// 2] a set of all explored game states.
	exploredGameStates map[GameStateCanonicalForm]bool

	// 3] a stack of game states.
	gameStateStack []GameState
}

func NewDFSGameSolver() DFSGameSolver {
	return DFSGameSolver{
		gameStateForGivenCanonicalForm:           make(map[GameStateCanonicalForm]GameState, 0),
		gameStateTransitionForGivenCanonicalForm: make(map[GameStateCanonicalForm]GameStateTransition, 0),
		exploredGameStates:                       make(map[GameStateCanonicalForm]bool, 0),
		gameStateStack:                           make([]GameState, 0),
	}
}

// Solve solves the ballsort puzzle given the initial game state.
// If there is a solution, it returns a sequence of state transitions required.
// If there is no solution, it returns nil.
//
// In any case, it returns some stats.
func (s *DFSGameSolver) Solve(cfg GameConfig, startingState GameState) GameSolution[DFSSearchStats] {

	s.visitInitialGameState(cfg, startingState)

	for s.unsolved() {
		maybeSolution := s.exploreNextGameState(cfg)
		if maybeSolution != nil {
			return *maybeSolution
		}
	}

	// no solution.
	return GameSolution[DFSSearchStats]{
		Stats: s.stats(),
	}

}

func (s *DFSGameSolver) visitInitialGameState(cfg GameConfig, gameState GameState) {
	s.visitGameStateViaTransition(cfg, gameState, nil)
}

func (s *DFSGameSolver) visitGameStateViaTransition(cfg GameConfig, gameState GameState, gameStateTransition *GameStateTransition) {
	s.gameStateStack = append(s.gameStateStack, gameState)

	gameStateCanonicalForm := gameState.CanonicalForm(cfg)

	s.gameStateForGivenCanonicalForm[gameStateCanonicalForm] = gameState
	if gameStateTransition != nil {
		s.gameStateTransitionForGivenCanonicalForm[gameStateCanonicalForm] = *gameStateTransition
	}
}

func (s DFSGameSolver) unsolved() bool {
	return len(s.gameStateStack) > 0
}

func (s *DFSGameSolver) exploreNextGameState(cfg GameConfig) *GameSolution[DFSSearchStats] {
	gameState := s.gameStateStack[len(s.gameStateStack)-1]
	s.gameStateStack = s.gameStateStack[:len(s.gameStateStack)-1]

	// If the game state is terminal, then the game has been solved.
	if gameState.IsTerminal(cfg) {
		return &GameSolution[DFSSearchStats]{
			Transitions: s.stitchGameStateTransitions(cfg, gameState),
			Stats:       s.stats(),
		}
	}

	gameStateCanonicalForm := gameState.CanonicalForm(cfg)

	// If the game state has been explored already, no need to do it again.
	if _, ok := s.exploredGameStates[gameStateCanonicalForm]; ok {
		return nil
	}

	// fmt.Printf("Exploring %s, num left %d\n", gameState.CanonicalForm(cfg), len(s.gameStateStack))

	// Then, explore the game state.
	gameStateTransitions := s.validGameStateTransitions(cfg, gameState)
	for _, gameStateTransition := range gameStateTransitions {
		s.visitGameStateViaTransition(cfg, gameStateTransition.ToGameState, &gameStateTransition)
	}

	// Mark the game state as explored.
	s.exploredGameStates[gameStateCanonicalForm] = true

	return nil // game not solved yet.
}

// validGameStateTransitions gets a set of transitions that we can make
// by moving a certain number of balls from one container to another in a given game state.
//
// For performance reasons, does two things.
//  1. reuses a known game state if one of the neighboring states
//     is equivalent to that known game state. And,
//  2. if a game state has already been explored, then doesn't include it
//     in the output.
func (s DFSGameSolver) validGameStateTransitions(cfg GameConfig, startingGameState GameState) []GameStateTransition {

	neighboringGameStateTransitions := make([]GameStateTransition, 0)

	startingGameStateCanonicalForm := startingGameState.CanonicalForm(cfg)

	for fromContainerIdx := 0; fromContainerIdx < cfg.NumContainers; fromContainerIdx++ {
		for toContainerIdx := 0; toContainerIdx < cfg.NumContainers; toContainerIdx++ {

			// Cannot transfer balls from a container to itself.
			if fromContainerIdx == toContainerIdx {
				continue
			}

			// Cannot transfer balls from an empty container.
			fromContainer := startingGameState.Containers[fromContainerIdx]
			numBallsInFromContainer := fromContainer.NumBalls()
			if numBallsInFromContainer == 0 {
				continue
			}

			// Cannot transfer balls of different color to non-empty container.
			toContainer := startingGameState.Containers[toContainerIdx]
			numBallsInToContainer := toContainer.NumBalls()
			if numBallsInToContainer == cfg.MaxNumBallsPerContainer {
				continue
			}
			if numBallsInToContainer > 0 && (fromContainer[numBallsInFromContainer-1] != toContainer[numBallsInToContainer-1]) {
				continue
			}

			// Cannot transfer balls to a container that doesn't have enough capacity.
			numBallsToMove := 1 // we know we can transfer at least one, since the from-container isn't empty.
			for j := numBallsInFromContainer - 2; j >= 0; j-- {
				if fromContainer[j] != fromContainer[numBallsInFromContainer-1] {
					break
				}
				numBallsToMove++
			}

			if numBallsToMove > (cfg.MaxNumBallsPerContainer - numBallsInToContainer) {
				continue
			}

			// No point transferring all balls from one container to an empty container.
			if numBallsToMove == numBallsInFromContainer && numBallsInToContainer == 0 {
				continue
			}

			// Found possible transition. Create correspond game state object.
			neighboringGameState := startingGameState.CloneWithBallsMoved(fromContainerIdx, toContainerIdx, numBallsToMove)

			// Get the canonical form of the neighboring game state.
			neighboringGameStateCanonicalForm := neighboringGameState.CanonicalForm(cfg)

			// Cannot consider a neighboring game state that is equivalent to the current state as a valid transition.
			if neighboringGameStateCanonicalForm == startingGameStateCanonicalForm {
				continue
			}

			// If the neighboring game state or its equivalent have already been visited before,
			// then we use the visited game state.
			neighboringGameStateEquivalent, found := s.gameStateForGivenCanonicalForm[neighboringGameStateCanonicalForm]
			if found {
				neighboringGameState = neighboringGameStateEquivalent
			}

			// If the neighboring gare state has been explored, then we ignore it.
			if _, alreadyExplored := s.exploredGameStates[neighboringGameStateCanonicalForm]; alreadyExplored {
				continue
			}

			neighboringGameStateTransitions = append(neighboringGameStateTransitions, GameStateTransition{
				FromContainerIdx: fromContainerIdx,
				ToContainerIdx:   toContainerIdx,
				NumBalls:         numBallsToMove,

				FromGameState: startingGameState,
				ToGameState:   neighboringGameState,
			})
		}
	}

	return neighboringGameStateTransitions
}

func (s DFSGameSolver) stitchGameStateTransitions(cfg GameConfig, terminalGameState GameState) []GameStateTransition {

	gameStateTransitions := make([]GameStateTransition, 0)

	gameState := terminalGameState
	for {
		gameStateCanonicalForm := gameState.CanonicalForm(cfg)

		gameStateTransition, ok := s.gameStateTransitionForGivenCanonicalForm[gameStateCanonicalForm]
		if !ok {
			// We have found all transitions.
			break
		}

		// fmt.Printf("Stiching: From state %s to %s \n", gameStateCanonicalForm, gameStateTransition.ToGameState.CanonicalForm(cfg))

		gameStateTransitions = append(gameStateTransitions, gameStateTransition)
		gameState = gameStateTransition.FromGameState
	}

	// Now, we need to reverse the set of transitions, because so far they were backwards from
	// a terminal state to the initial starting state, but we need it the other way around.
	slices.Reverse(gameStateTransitions)
	return gameStateTransitions
}

func (s DFSGameSolver) stats() DFSSearchStats {
	return DFSSearchStats{
		NumVisitedGameStates:  len(s.gameStateForGivenCanonicalForm),
		NumExploredGameStates: len(s.exploredGameStates),
	}
}

//////////////////////////////////////////
// Data structures for input parsing
//////////////////////////////////////////

type RawGameInput struct {
	GameConfig GameConfig

	GameState RawGameState
}

type RawContainer []string

type RawGameState struct {
	Containers []RawContainer
}

func (rgs RawGameState) GetGameState(cfg GameConfig) GameState {

	colorsMap := make(map[string]int, 0)
	for i, color := range cfg.Colors {
		colorsMap[color] = i + 1
	}

	numContainers := len(rgs.Containers)

	clonedContainers := make([]Container, 0, numContainers)

	for i := 0; i < numContainers; i++ {

		container := rgs.Containers[i]

		clonedContainer := make([]int, len(container))
		for j := 0; j < len(container); j++ {
			clonedContainer[j] = colorsMap[container[j]]
		}

		clonedContainers = append(clonedContainers, clonedContainer)
	}

	return GameState{
		Containers: clonedContainers,
	}
}

//////////////////////////////////////////
// Main
//////////////////////////////////////////

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

	// Convert from raw input to internal format to pass to the solver.
	gameConfig := gi.GameConfig
	initialGameState := gi.GameState.GetGameState(gameConfig)

	dfsGameSolver := NewDFSGameSolver()
	solution := dfsGameSolver.Solve(gameConfig, initialGameState)

	fmt.Printf(
		"Search stats: num visited states %d, num explored states %d\n", solution.Stats.NumVisitedGameStates, solution.Stats.NumExploredGameStates)

	if solution.Transitions == nil {
		fmt.Println("No solution found")
		return
	}

	fmt.Printf("Solution found with %d steps\n", len(solution.Transitions))
	for i := 0; i < len(solution.Transitions); i++ {
		gameTransition := solution.Transitions[i]
		fmt.Printf("Step %3d: Move %d balls from %3d to %3d container\n", i+1, gameTransition.NumBalls, gameTransition.FromContainerIdx+1, gameTransition.ToContainerIdx+1)
	}
}
