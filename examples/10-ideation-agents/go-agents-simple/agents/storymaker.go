// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Simplified Ideation Agents

package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

// StorymakerDefinition implements AgentDefinition for the Storymaker agent
type StorymakerDefinition struct {
	*agent.BaseAgentDefinition
}

// NewStorymakerDefinition creates a new Storymaker agent definition
func NewStorymakerDefinition() *StorymakerDefinition {
	config := &agent.AgentConfig{
		Name:        "Story Creator",
		Type:        agent.AgentTypeChat,
		Model:       "gemma3:1b",
		Provider:    "ollama",
		Temperature: 0.7,
		MaxTokens:   500,
		SystemPrompt: `You are a creative storyteller specializing in crafting engaging narratives about future habitat scenarios and sustainable living in 2035.

Your expertise includes:
- Creating compelling stories about future living spaces
- Integrating sustainability themes into narratives
- Developing character-driven scenarios around habitat design
- Weaving together technical requirements with human experiences
- Inspiring audiences through imaginative yet plausible future scenarios

Create stories that help people envision themselves living in sustainable habitats.
Make complex sustainability concepts accessible through storytelling.`,
	}

	definition := &StorymakerDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
	}

	// Set comprehensive schema metadata for auto-validation
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"story_prompt": map[string]interface{}{
				"type":        "string",
				"description": "Prompt or theme for the story",
				"minLength":   10,
				"maxLength":   1000,
			},
			"setting": map[string]interface{}{
				"type":        "object",
				"description": "Optional story setting details",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "Geographic location or environment",
						"maxLength":   100,
					},
					"time_period": map[string]interface{}{
						"type":        "string",
						"description": "Time period for the story",
						"default":     "2035",
						"maxLength":   50,
					},
					"habitat_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of habitat featured",
						"enum":        []interface{}{"urban", "suburban", "rural", "off-grid", "floating", "underground", "vertical"},
					},
				},
			},
			"characters": map[string]interface{}{
				"type":        "array",
				"description": "Optional character descriptions",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":      "string",
							"maxLength": 50,
						},
						"role": map[string]interface{}{
							"type":      "string",
							"maxLength": 100,
						},
						"background": map[string]interface{}{
							"type":      "string",
							"maxLength": 200,
						},
					},
				},
			},
		},
		"required":    []string{"story_prompt"},
		"description": "Input schema for Storymaker agent",
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Story title",
				"minLength":   5,
				"maxLength":   100,
			},
			"story": map[string]interface{}{
				"type":        "string",
				"description": "The complete story narrative",
				"minLength":   200,
				"maxLength":   5000,
			},
			"themes": map[string]interface{}{
				"type":        "array",
				"description": "Key themes explored in the story",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 3,
					"maxLength": 50,
				},
			},
			"sustainability_features": map[string]interface{}{
				"type":        "array",
				"description": "Sustainability features highlighted in the story",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 10,
					"maxLength": 150,
				},
			},
			"moral": map[string]interface{}{
				"type":        "string",
				"description": "Key takeaway or moral of the story",
				"minLength":   20,
				"maxLength":   300,
			},
			"target_audience": map[string]interface{}{
				"type":        "string",
				"description": "Intended audience for the story",
				"enum":        []interface{}{"children", "young_adults", "adults", "professionals", "general"},
			},
			"genre": map[string]interface{}{
				"type":        "string",
				"description": "Story genre",
				"enum":        []interface{}{"science_fiction", "slice_of_life", "adventure", "drama", "educational", "utopian"},
			},
		},
		"required":    []string{"title", "story", "themes", "moral"},
		"description": "Output schema for Storymaker agent",
	})

	definition.SetMetadata("description", "Creates engaging narratives about future habitat scenarios")
	definition.SetMetadata("tags", []string{"storytelling", "narrative", "futures", "sustainability"})

	return definition
}

// GetStorymakerConfig returns the configuration for backward compatibility
func GetStorymakerConfig() *agent.AgentConfig {
	return NewStorymakerDefinition().GetConfig()
}

// CreateAgent creates a Storymaker agent with custom graph workflow
func (s *StorymakerDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := s.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build custom graph for Storymaker workflow
	graph := core.NewGraph("storymaker-workflow")
	graph.AddNode("plan_narrative", "Plan Narrative", s.planNarrativeNode)
	graph.AddNode("generate_story", "Generate Story", s.generateStoryNode)

	graph.SetStartNode("plan_narrative")
	graph.AddEdge("plan_narrative", "generate_story", nil)
	graph.AddEndNode("generate_story")

	// Set the custom graph on the agent
	baseAgent.SetGraph(graph)

	return baseAgent, nil
}

// planNarrativeNode handles story planning and structure
func (s *StorymakerDefinition) planNarrativeNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

	var storyPrompt string
	switch v := inputData.(type) {
	case string:
		storyPrompt = v
	case map[string]interface{}:
		if prompt, ok := v["story_prompt"].(string); ok {
			storyPrompt = prompt
		} else if message, ok := v["message"].(string); ok {
			storyPrompt = message
		} else {
			return nil, fmt.Errorf("no story_prompt or message field in input")
		}
	default:
		return nil, fmt.Errorf("invalid input type")
	}

	// Analyze story prompt to determine narrative theme and structure
	lowerPrompt := strings.ToLower(storyPrompt)
	var narrativeTheme string
	var storyStructure string

	if strings.Contains(lowerPrompt, "energy") || strings.Contains(lowerPrompt, "solar") || strings.Contains(lowerPrompt, "power") {
		narrativeTheme = "energy_innovation"
		storyStructure = "Tech Discovery → Challenge → Innovation → Transformation"
	} else if strings.Contains(lowerPrompt, "water") || strings.Contains(lowerPrompt, "rain") || strings.Contains(lowerPrompt, "ocean") {
		narrativeTheme = "water_harmony"
		storyStructure = "Scarcity → Discovery → Integration → Abundance"
	} else if strings.Contains(lowerPrompt, "urban") || strings.Contains(lowerPrompt, "city") || strings.Contains(lowerPrompt, "vertical") {
		narrativeTheme = "urban_evolution"
		storyStructure = "Density Crisis → Vertical Solution → Community → New Urban Life"
	} else if strings.Contains(lowerPrompt, "nature") || strings.Contains(lowerPrompt, "forest") || strings.Contains(lowerPrompt, "bio") {
		narrativeTheme = "biophilic_connection"
		storyStructure = "Disconnection → Discovery → Integration → Symbiosis"
	} else if strings.Contains(lowerPrompt, "future") || strings.Contains(lowerPrompt, "2035") || strings.Contains(lowerPrompt, "tomorrow") {
		narrativeTheme = "future_vision"
		storyStructure = "Present Challenge → Time Jump → Future Solution → Hope"
	} else {
		narrativeTheme = "sustainable_living"
		storyStructure = "Problem → Innovation → Implementation → Impact"
	}

	plan := fmt.Sprintf(`📖 **Narrative Plan for: "%s"**

**Theme:** %s story
**Structure:** %s
**Setting:** Sustainable habitat in 2035
**Focus:** Human experience within innovative environmental solutions
**Tone:** Inspiring yet realistic, showcasing technology serving humanity and nature`, storyPrompt, narrativeTheme, storyStructure)

	state.Set("narrative_plan", plan)
	state.Set("story_prompt", storyPrompt)
	state.Set("narrative_theme", narrativeTheme)
	return state, nil
}

// generateStoryNode creates the complete story narrative
func (s *StorymakerDefinition) generateStoryNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	storyPrompt, _ := state.Get("story_prompt")
	narrativeTheme, _ := state.Get("narrative_theme")

	var title string
	var story string
	var themes []string
	var sustainabilityFeatures []string
	var moral string
	var targetAudience string
	var genre string

	// Generate contextual stories based on narrative theme
	switch narrativeTheme {
	case "energy_innovation":
		title = "The Solar Symphony"
		story = fmt.Sprintf(`⚡ **The Solar Symphony**

Dr. Elena Rodriguez stepped into her new energy-positive habitat for the first time, and the building seemed to greet her. "%s" - that's exactly what she'd been searching for since the energy crisis of 2033.

The transparent photovoltaic walls shimmered like liquid crystal, each panel dancing with captured sunlight. As Elena moved through her new home, kinetic sensors in the flooring harvested energy from her footsteps, adding to the building's power reserves. The AI system, Aurora, whispered gentle updates: "Good morning, Elena. Today's energy generation: 340%% of consumption. Excess power contributed to community grid."

But the real magic happened at sunset. The building's battery walls, recycled from old electric vehicles, began to glow with stored energy. The entire structure became a beacon of sustainability, visible from across the valley. Elena realized she wasn't just living in a home - she was part of a living power grid that breathed with the rhythm of sun and wind.

When the neighborhood children gathered to watch the evening light show, Elena knew this wasn't just about energy efficiency. This was about creating a world where technology and nature performed in perfect harmony, composing a symphony of sustainable living for generations to come.`, storyPrompt)
		themes = []string{"renewable energy", "community cooperation", "technological harmony", "environmental stewardship", "innovation"}
		sustainabilityFeatures = []string{"building-integrated photovoltaics", "kinetic energy harvesting", "recycled battery storage", "community energy sharing", "smart grid integration"}
		moral = "True energy independence comes not from individual consumption, but from creating systems that generate abundance for entire communities."
		targetAudience = "adults"
		genre = "science_fiction"

	case "water_harmony":
		title = "Rivers in the Sky"
		story = fmt.Sprintf(`💧 **Rivers in the Sky**

Marcus had always loved rain, but living in the Atmospheric Water Gardens changed everything. "%s" - his grandmother's words echoed as he watched the morning mist transform into drinking water before his eyes.

The habitat's living walls breathed with the humidity, extracting precious moisture from the air through bio-engineered moss that his neighbor Dr. Chen had perfected. Water flowed through transparent tubes like arteries, showing the complete cycle from collection to purification to reuse. The aquaponic gardens on every level turned water into food, while the overflow cascaded down sculptural channels that served as both art and infrastructure.

But it was during the first major storm that Marcus truly understood the system's genius. Instead of flooding, the building embraced the deluge. Rain danced through collection channels, filtered through living walls, and fed the community's growing gardens. The excess flowed into underground reserves that would sustain them through dry seasons.

Standing on his balcony, watching water cycle through his neighborhood like a natural watershed, Marcus realized they hadn't just solved water scarcity - they'd made every drop a celebration of life itself.`, storyPrompt)
		themes = []string{"water conservation", "bio-engineering", "cyclical systems", "community resilience", "natural integration"}
		sustainabilityFeatures = []string{"atmospheric water generation", "bio-engineered filtration", "aquaponic food systems", "rainwater sculpture", "underground reserves"}
		moral = "Water scarcity transforms into abundance when we learn to dance with natural cycles rather than fight them."
		targetAudience = "general"
		genre = "slice_of_life"

	case "urban_evolution":
		title = "Vertical Village"
		story = fmt.Sprintf(`🏙️ **Vertical Village**

Zara pressed her palm against the elevator button and smiled as it powered up from her bioelectric touch. "%s" - she thought, remembering when cities felt like concrete prisons rather than living communities.

Her apartment was part of a 50-story vertical village where each floor specialized in different aspects of community life. Floor 15 housed the shared workshops where neighbors built furniture from recycled materials. Floor 23 contained the sky gardens where children learned about food systems. Floor 31 was the community center where decisions were made collectively through digital democracy platforms.

The magic happened in the connecting sky bridges between buildings. What started as transportation quickly evolved into aerial neighborhoods, with markets, cafes, and meeting spaces suspended between the towers. Zara's favorite was the Bridge of Stories on Level 25, where elders shared wisdom while children played in gravity-defying gardens.

From her balcony that evening, watching the city pulse with life across dozens of connected towers, Zara understood they hadn't just solved urban density. They'd rediscovered what it meant to be a village - only now their village reached toward the sky, and their neighbors numbered in the thousands.`, storyPrompt)
		themes = []string{"vertical living", "community connection", "shared resources", "urban sustainability", "collective decision-making"}
		sustainabilityFeatures = []string{"bioelectric interfaces", "sky bridge communities", "vertical gardens", "shared workshop spaces", "digital democracy systems"}
		moral = "Cities become truly sustainable when they prioritize human connection alongside efficient land use."
		targetAudience = "young_adults"
		genre = "utopian"

	case "biophilic_connection":
		title = "The Growing House"
		story = fmt.Sprintf(`🌱 **The Growing House**

When Jamie first moved into the mycelium house, the walls were still learning her scent. "%s" - her architect had explained, but experiencing a living building was something entirely different.

The house grew around her daily rhythms. Mycelium networks in the walls sensed her sleep patterns and adjusted air purification accordingly. When she felt stressed, bioluminescent panels responded with calming blue light. The structure itself was part of the forest ecosystem - tree roots intertwined with foundation networks, and the roof supported a thriving ecosystem of birds, insects, and small mammals.

What amazed Jamie most was how the house healed itself. When winter storms damaged the living walls, the mycelium regrew stronger than before. Waste became food for the structure's biological processes. Even the air felt different - not just filtered, but actively enhanced by the breathing walls.

Six months later, when friends visited Jamie's truly living home, they watched in wonder as the walls literally pulsed with life, the windows dimmed naturally with the sunset, and the house seemed to welcome them with a gentle, earthy fragrance. Jamie realized she wasn't just living in harmony with nature - she had become part of it.`, storyPrompt)
		themes = []string{"living architecture", "symbiotic relationships", "adaptive systems", "natural integration", "regenerative design"}
		sustainabilityFeatures = []string{"mycelium construction", "bioluminescent lighting", "self-healing materials", "integrated ecosystems", "biological air purification"}
		moral = "The future of sustainable living lies not in conquering nature, but in becoming indistinguishable from it."
		targetAudience = "adults"
		genre = "science_fiction"

	case "future_vision":
		title = "Letters from 2035"
		story = fmt.Sprintf(`🔮 **Letters from 2035**

"Dear Past Self," Maria typed into her holographic journal. "%s" - you probably can't imagine how we solved this, but let me tell you about a typical day in 2035.

I wake up as my bedroom walls gradually brighten, mimicking sunrise even on cloudy days. The air tastes clean - not just filtered, but actively enriched by the building's living wall systems. My coffee comes from beans grown in the vertical farms three floors below, watered by yesterday's shower.

The commute that used to stress you? Gone. My office is part of a network of shared spaces accessible by sky tram. When I need collaboration, I meet colleagues in person. When I need focus, I work from neighborhood pods designed for deep thinking.

But here's what you really need to know: we didn't sacrifice comfort for sustainability. We discovered that the most sustainable solutions were also the most beautiful, the most connected, the most human. The technology you're worried about? It faded into the background, making life simpler, not more complex.

Your choices matter more than you know. Every sustainable decision you make creates ripples that reach us here in 2035. Keep believing in the future - from here, I can tell you it's even more beautiful than you dare to dream."`, storyPrompt)
		themes = []string{"intergenerational hope", "technological integration", "sustainable progress", "community evolution", "optimistic futures"}
		sustainabilityFeatures = []string{"circadian lighting systems", "building-integrated agriculture", "shared workspace networks", "atmospheric purification", "background technology"}
		moral = "The future is not something that happens to us, but something we create through our daily choices and collective vision."
		targetAudience = "general"
		genre = "educational"

	default: // sustainable_living
		title = "The Neighborhood Revolution"
		story = fmt.Sprintf(`🏘️ **The Neighborhood Revolution**

It started with Sarah's simple request: "%s" - and somehow that conversation at the neighborhood meeting changed everything.

Within two years, their suburban block had transformed into something unrecognizable. Solar canopies stretched between houses, sharing energy across property lines. Rain gardens connected backyards in a continuous watershed. The community workshop in the Chen family's garage had grown into a tool library that served twelve neighborhoods.

But the real revolution was social. Block parties became planning sessions. Children moved freely between houses, learning different skills from each family. Elderly neighbors became teachers and storytellers. The artificial barriers between properties dissolved into collaborative spaces.

Sarah watched from her front porch as her neighbors tested the new community-scale compost system. Mrs. Rodriguez was explaining soil chemistry to a group of fascinated eight-year-olds. The Martinez family was harvesting vegetables from the shared food forest that had replaced the old lawns. And in the distance, she could see similar transformations spreading to other neighborhoods like a gentle, green wave.

"We didn't just make our neighborhood sustainable," Sarah realized. "We remembered how to be neighbors."`, storyPrompt)
		themes = []string{"community transformation", "resource sharing", "local resilience", "social sustainability", "collaborative living"}
		sustainabilityFeatures = []string{"shared energy systems", "watershed management", "tool libraries", "community composting", "food forest integration"}
		moral = "Sustainability is not just about technology and efficiency - it's about rediscovering the strength that comes from genuine community."
		targetAudience = "general"
		genre = "slice_of_life"
	}

	// Structure the output according to schema
	output := map[string]interface{}{
		"title":                   title,
		"story":                   story,
		"themes":                  themes,
		"sustainability_features": sustainabilityFeatures,
		"moral":                   moral,
		"target_audience":         targetAudience,
		"genre":                   genre,
	}

	state.Set("output", output)
	return state, nil
}
