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

// DesignerDefinition implements AgentDefinition for the Designer agent
type DesignerDefinition struct {
	*agent.BaseAgentDefinition
}

// NewDesignerDefinition creates a new Designer agent definition
func NewDesignerDefinition() *DesignerDefinition {
	config := &agent.AgentConfig{
		Name:        "Visual Designer",
		Type:        agent.AgentTypeChat,
		Model:       "gemma3:1b",
		Provider:    "ollama",
		Temperature: 0.7,
		MaxTokens:   500,
		SystemPrompt: `You are a visionary creative designer specializing in sustainable architecture and habitat design for the year 2035.

Your expertise includes:
- Eco-friendly materials and construction techniques
- Energy-efficient design principles
- Integration with natural environments
- Smart home technology integration
- Sustainable living solutions

Always respond in a structured format with design descriptions, materials, and style categorization.
Provide detailed, imaginative descriptions that inspire and inform.`,
	}

	definition := &DesignerDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
	}

	// Set comprehensive schema metadata for auto-validation
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Design request or description",
				"minLength":   1,
				"maxLength":   1000,
				"pattern":     "^.+$",
			},
		},
		"required":    []string{"message"},
		"description": "Input schema for Designer agent",
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Design description",
				"minLength":   10,
			},
			"image_data": map[string]interface{}{
				"type":        "string",
				"description": "Base64 encoded image data",
				"pattern":     "^data:image/(png|jpeg|jpg|gif|webp);base64,[A-Za-z0-9+/]+={0,2}$",
			},
			"materials": map[string]interface{}{
				"type":        "array",
				"description": "List of materials used",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"style": map[string]interface{}{
				"type":        "string",
				"description": "Design style category",
				"enum":        []interface{}{"modern", "minimalist", "futuristic", "organic", "industrial"},
			},
		},
		"required":    []string{"description", "image_data", "style"},
		"description": "Output schema for Designer agent",
	})

	definition.SetMetadata("description", "Creates visual designs and concepts for habitat spaces")
	definition.SetMetadata("tags", []string{"design", "architecture", "sustainability", "creativity"})

	return definition
}

// CreateAgent creates a Designer agent with custom graph workflow
func (d *DesignerDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := d.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build custom graph for Designer workflow
	graph := core.NewGraph("designer-workflow")
	graph.AddNode("design", "Generate Design", d.designNode)
	graph.SetStartNode("design")
	graph.AddEndNode("design")

	// Set the custom graph on the agent
	baseAgent.SetGraph(graph)

	return baseAgent, nil
}

// designNode handles the design generation logic with custom workflow
func (d *DesignerDefinition) designNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Get input from state
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

	// Extract message from input
	var message string
	switch v := inputData.(type) {
	case string:
		message = v
	case map[string]interface{}:
		if msg, ok := v["message"].(string); ok {
			message = msg
		} else {
			return nil, fmt.Errorf("no message field in input")
		}
	default:
		return nil, fmt.Errorf("invalid input type")
	}

	// Analyze input to provide contextual design responses
	var response string
	var materials []string
	var style string

	// Convert to lowercase for pattern matching
	lowerMessage := strings.ToLower(message)

	// Contextual design generation based on input content
	if strings.Contains(lowerMessage, "energy") || strings.Contains(lowerMessage, "solar") || strings.Contains(lowerMessage, "renewable") {
		response = fmt.Sprintf(`🔋 **Energy-Optimized Habitat Design 2035 - "%s"**

**Design Concept:**
A revolutionary energy-positive habitat that produces more power than it consumes. This design integrates cutting-edge renewable energy systems with intelligent consumption management.

**Energy Features:**
• Transparent photovoltaic glass panels covering 80%% of exterior surfaces
• Vertical wind turbines integrated into building corners
• Geothermal heat pumps for temperature regulation
• Kinetic energy harvesting from foot traffic and door movements
• Battery walls using recycled EV batteries for energy storage
• Smart energy distribution AI optimizing usage patterns

**Architectural Style:** Clean lines with high-tech integration
**Energy Performance:** 150%% energy positive - surplus feeds community grid
**Innovation:** Building-integrated photovoltaics (BIPV) with 45%% efficiency

This design transforms the habitat into a living power station while maintaining aesthetic appeal and comfort.`, message)
		materials = []string{"photovoltaic glass", "carbon fiber wind turbines", "recycled battery cells", "aerogel insulation", "smart conductors"}
		style = "futuristic"

	} else if strings.Contains(lowerMessage, "water") || strings.Contains(lowerMessage, "rain") || strings.Contains(lowerMessage, "hydro") {
		response = fmt.Sprintf(`💧 **Water-Integrated Habitat Design 2035 - "%s"**

**Design Concept:**
A water-conscious habitat featuring advanced water management systems that collect, purify, and recycle every drop. The design celebrates water as both a resource and aesthetic element.

**Water Management Features:**
• Sculptural rain collection channels integrated into roof design
• Living water walls with aquatic plants for natural filtration
• Atmospheric water generators extracting moisture from air
• Greywater recycling systems with transparent viewing tubes
• Permeable building materials allowing controlled water flow
• Smart irrigation systems for integrated food gardens

**Architectural Style:** Flowing, organic forms mimicking water movement
**Water Efficiency:** 95%% water recycling rate with zero waste
**Innovation:** Bio-mimetic water collection inspired by desert beetles

This design creates a symbiotic relationship between inhabitants and the water cycle.`, message)
		materials = []string{"permeable bio-concrete", "aquatic filtration systems", "water-reactive ceramics", "hydrophilic surfaces", "living algae panels"}
		style = "organic"

	} else if strings.Contains(lowerMessage, "urban") || strings.Contains(lowerMessage, "city") || strings.Contains(lowerMessage, "vertical") {
		response = fmt.Sprintf(`🏙️ **Vertical Urban Habitat Design 2035 - "%s"**

**Design Concept:**
A space-efficient vertical habitat optimized for dense urban environments. This design maximizes living space while creating green corridors that connect to urban nature networks.

**Urban Integration Features:**
• Modular stacking system allowing vertical expansion
• Sky bridges connecting to adjacent buildings and transport
• Vertical gardens creating living facades that filter air
• Flexible interior spaces with moveable walls and furniture
• Community spaces on alternating floors for social interaction
• Drone delivery ports integrated into balcony design

**Architectural Style:** Sleek verticality with green integration
**Space Efficiency:** 40%% more usable space than traditional apartments
**Innovation:** AI-optimized space reconfiguration responding to daily needs

This design proves that urban density and quality of life can coexist harmoniously.`, message)
		materials = []string{"lightweight steel frames", "living wall substrates", "smart glass partitions", "carbon fiber connectors", "urban-adapted plants"}
		style = "modern"

	} else if strings.Contains(lowerMessage, "natural") || strings.Contains(lowerMessage, "forest") || strings.Contains(lowerMessage, "bio") || strings.Contains(lowerMessage, "tree") {
		response = fmt.Sprintf(`🌲 **Bio-Integrated Habitat Design 2035 - "%s"**

**Design Concept:**
A living habitat that blurs the boundary between architecture and nature. This design grows with the ecosystem, using biological processes as functional building components.

**Bio-Integration Features:**
• Mycelium-based structural walls that grow and self-repair
• Living roof ecosystem supporting local biodiversity
• Tree-integrated support pillars using actual mature trees
• Bioluminescent lighting from genetically modified plants
• Composting waste systems feeding building's living components
• Air purification through integrated plant respiratory systems

**Architectural Style:** Biomimetic forms following natural growth patterns
**Environmental Impact:** Carbon negative - absorbs more CO2 than construction produced
**Innovation:** Self-healing bio-materials that adapt to environmental stress

This design represents true symbiosis between human habitation and natural ecosystems.`, message)
		materials = []string{"mycelium composites", "living wood structures", "bio-luminescent plants", "compostable polymers", "symbiotic organisms"}
		style = "organic"

	} else if strings.Contains(lowerMessage, "luxury") || strings.Contains(lowerMessage, "premium") || strings.Contains(lowerMessage, "elegant") {
		response = fmt.Sprintf(`✨ **Luxury Sustainable Habitat Design 2035 - "%s"**

**Design Concept:**
Where opulence meets responsibility - a habitat that proves luxury and sustainability are not mutually exclusive. Every premium feature incorporates environmental consciousness.

**Luxury Sustainability Features:**
• Rare earth-free smart materials with premium aesthetics
• Artisanal bio-fabricated surfaces unique to each installation
• Climate-controlled micro-environments for different activities
• Holographic entertainment systems powered by renewable energy
• Automated gourmet hydroponic gardens with exotic varieties
• Personalized air composition systems for optimal health

**Architectural Style:** Minimalist luxury with subtle technological integration
**Comfort Level:** Five-star hotel comfort with zero environmental impact
**Innovation:** Lab-grown diamond windows providing superior insulation

This design demonstrates that the future of luxury lies in harmony with nature.`, message)
		materials = []string{"lab-grown diamond surfaces", "artisanal bio-fabrics", "rare-wood alternatives", "precious metal recycling", "smart textile integration"}
		style = "minimalist"

	} else if len(message) > 100 {
		// Long, detailed input - provide comprehensive design analysis
		response = fmt.Sprintf(`🏗️ **Comprehensive Habitat Design Analysis 2035 - "%s"**

**Design Concept:**
Based on your detailed requirements, I've developed a multi-faceted habitat design that addresses all aspects of sustainable living while maintaining human comfort and aesthetic appeal.

**Integrated Systems Analysis:**
• Holistic approach combining energy, water, waste, and social needs
• Adaptive design responding to seasonal and lifestyle changes
• Resilient systems with multiple redundancies for reliability
• Community integration features supporting social sustainability
• Future-proofing with upgradeable modular components
• Health-optimized environments using biophilic design principles

**Technical Innovation:**
This design synthesizes multiple advanced technologies:
- AI-driven environmental optimization
- Circular economy material flows
- Passive house standards exceeded by 200%%
- Social spaces designed using behavioral psychology research

**Architectural Style:** Integrated complexity with human-centered simplicity
**Sustainability Rating:** Regenerative - leaves environment better than before
**Innovation:** First truly net-positive habitat design

This comprehensive approach ensures your habitat becomes a model for sustainable living.`, message)
		materials = []string{"multi-functional composites", "regenerative bio-materials", "adaptive smart systems", "community-sourced elements", "future-compatible interfaces"}
		style = "futuristic"

	} else {
		// General/greeting responses
		response = fmt.Sprintf(`🏠 **Foundational Sustainable Habitat Design 2035 - "%s"**

**Design Concept:**
A thoughtfully designed habitat that establishes the foundation for sustainable living in 2035. This design focuses on proven technologies and timeless principles while preparing for future innovations.

**Core Sustainable Features:**
• Passive solar design optimizing natural light and heat
• High-performance insulation reducing energy needs by 80%%
• Smart home systems learning and adapting to occupant preferences
• Integrated food production spaces for self-sufficiency
• Rainwater collection and greywater recycling systems
• Flexible spaces adapting to changing life circumstances

**Architectural Style:** Contemporary sustainability with classic comfort
**Performance:** Net-zero energy with potential for surplus generation
**Innovation:** Seamless integration of proven sustainable technologies

This design provides a solid foundation for sustainable living that can evolve with emerging technologies.`, message)
		materials = []string{"recycled steel frames", "hemp-crete insulation", "triple-glazed windows", "FSC-certified wood", "solar thermal collectors"}
		style = "modern"
	}

	// Create structured output
	output := map[string]interface{}{
		"description": response,
		"image_data":  d.createPlaceholderImage(),
		"materials":   materials,
		"style":       style,
	}

	state.Set("output", output)
	return state, nil
}

// generateWithLLM generates text using the LLM (simplified version)
func (d *DesignerDefinition) generateWithLLM(ctx context.Context, prompt string) (string, error) {
	// This is a simplified implementation for demo purposes
	// In a production environment, this would integrate with the agent's LLM provider
	return fmt.Sprintf("🏗️ Sustainable Habitat Design 2035\n\n%s\n\nThis design features cutting-edge sustainable architecture with integrated renewable energy systems, advanced water recycling, and smart automation for optimal living comfort while minimizing environmental impact.", prompt), nil
}

// createPlaceholderImage creates a base64 placeholder image
func (d *DesignerDefinition) createPlaceholderImage() string {
	// Simple 1x1 transparent PNG placeholder
	return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
}

// GetDesignerConfig returns the configuration for backward compatibility
func GetDesignerConfig() *agent.AgentConfig {
	return NewDesignerDefinition().GetConfig()
}
