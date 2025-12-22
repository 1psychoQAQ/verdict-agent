package artifact

import (
	"fmt"
	"time"

	"github.com/1psychoQAQ/verdict-agent/internal/agent"
	"github.com/1psychoQAQ/verdict-agent/internal/pipeline"
)

// ExampleGenerator_Generate demonstrates artifact generation
func ExampleGenerator_Generate() {
	g := NewGenerator()

	result := &pipeline.PipelineResult{
		Input: "Should I build a mobile app or web app first?",
		Verdict: &agent.VerdictOutput{
			Ruling:    "Build mobile app first",
			Rationale: "Mobile-first approach reaches users faster and provides better engagement",
			Rejected: []agent.RejectedOption{
				{
					Option: "web app",
					Reason: "requires more infrastructure setup and has lower user engagement",
				},
			},
			Ranking: []int{1, 2},
		},
		Execution: &agent.ExecutionOutput{
			MVPScope: []string{
				"User authentication",
				"Core feature implementation",
				"Basic UI/UX",
			},
			Phases: []agent.Phase{
				{
					Name: "Foundation",
					Tasks: []string{
						"Setup React Native project",
						"Configure build pipeline",
						"Setup authentication service",
					},
				},
				{
					Name: "Core Development",
					Tasks: []string{
						"Implement user login/signup",
						"Build main feature screens",
						"Add navigation",
					},
				},
				{
					Name: "Testing & Polish",
					Tasks: []string{
						"Write unit tests",
						"Conduct user testing",
						"Fix critical bugs",
					},
				},
			},
			DoneCriteria: []string{
				"App builds successfully on iOS and Android",
				"Users can authenticate and access core features",
				"No critical bugs in user flows",
			},
		},
		Duration: 5 * time.Minute,
	}

	artifacts, err := g.Generate(result)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== decision.json ===")
	fmt.Println(string(artifacts.DecisionJSON))
	fmt.Println()
	fmt.Println("=== todo.md ===")
	fmt.Println(string(artifacts.TodoMD))
}
