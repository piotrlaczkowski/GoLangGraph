// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

// StoryContext represents the current storytelling state
type StoryContext struct {
	SessionID       string           `json:"session_id"`
	StoriesCreated  int              `json:"stories_created"`
	PreferredGenres []string         `json:"preferred_genres"`
	CharacterBank   []StoryCharacter `json:"character_bank"`
	ThemeHistory    []string         `json:"theme_history"`
	LastCreation    time.Time        `json:"last_creation"`
	QualityScore    float64          `json:"quality_score"`
}

// StoryCharacter represents a character in the story
type StoryCharacter struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	Background string `json:"background"`
	Motivation string `json:"motivation"`
}

// StorySetting represents the setting for a story
type StorySetting struct {
	Location      string `json:"location"`
	TimePeriod    string `json:"time_period"`
	HabitatType   string `json:"habitat_type"`
	Environment   string `json:"environment"`
	SocialContext string `json:"social_context"`
}

// StoryOutput represents the complete story output
type StoryOutput struct {
	ID                     string                 `json:"id"`
	SessionID              string                 `json:"session_id"`
	Title                  string                 `json:"title"`
	Story                  string                 `json:"story"`
	Themes                 []string               `json:"themes"`
	SustainabilityFeatures []string               `json:"sustainability_features"`
	Moral                  string                 `json:"moral"`
	TargetAudience         string                 `json:"target_audience"`
	Genre                  string                 `json:"genre"`
	Characters             []StoryCharacter       `json:"characters"`
	Setting                StorySetting           `json:"setting"`
	WordCount              int                    `json:"word_count"`
	ReadingTime            string                 `json:"reading_time"`
	EducationalValue       string                 `json:"educational_value"`
	CreatedAt              time.Time              `json:"created_at"`
	Metadata               map[string]interface{} `json:"metadata"`
}

// StorymakingDefinition implements enhanced AgentDefinition for the Storymaker agent
type StorymakingDefinition struct {
	*agent.BaseAgentDefinition
	checkpointer persistence.Checkpointer
}

// NewStorymakerDefinition creates a new enhanced Storymaker agent definition
func NewStorymakerDefinition() *StorymakingDefinition {
	config := &agent.AgentConfig{
		Name:     "Story Creator",
		Type:     agent.AgentTypeChat,
		Model:    "gemma3:1b",
		Provider: "ollama",
		SystemPrompt: `You are a creative storyteller specializing in crafting engaging narratives about future habitat scenarios and sustainable living in 2035.

Your expertise includes:
- Creating compelling stories about future living spaces and sustainable communities
- Integrating sustainability themes into captivating narratives
- Developing character-driven scenarios around habitat design and environmental harmony
- Weaving together technical requirements with human experiences and emotions
- Inspiring audiences through imaginative yet plausible future scenarios
- Building educational narratives that promote sustainable living practices
- Adapting storytelling style for different target audiences
- Creating memorable characters that embody sustainable values

Create stories that help people envision themselves living in sustainable habitats.
Make complex sustainability concepts accessible and inspiring through storytelling.
Focus on human connections, emotional journeys, and practical sustainability solutions.`,
	}

	definition := &StorymakingDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
		checkpointer:        persistence.NewMemoryCheckpointer(),
	}

	// Set comprehensive schema metadata
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"story_prompt": map[string]interface{}{
				"type":        "string",
				"description": "Prompt or theme for the story",
				"minLength":   10,
				"maxLength":   2000,
			},
			"setting": map[string]interface{}{
				"type":        "object",
				"description": "Story setting details",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "Geographic location or environment",
						"maxLength":   200,
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
						"enum":        []interface{}{"urban", "suburban", "rural", "off-grid", "floating", "underground", "vertical", "community"},
					},
					"environment": map[string]interface{}{
						"type":        "string",
						"description": "Environmental context",
						"enum":        []interface{}{"forest", "coastal", "desert", "mountain", "arctic", "tropical", "urban", "mixed"},
					},
				},
			},
			"characters": map[string]interface{}{
				"type":        "array",
				"description": "Character descriptions",
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
							"maxLength": 300,
						},
						"motivation": map[string]interface{}{
							"type":      "string",
							"maxLength": 200,
						},
					},
				},
			},
			"target_audience": map[string]interface{}{
				"type":        "string",
				"description": "Intended audience for the story",
				"enum":        []interface{}{"children", "young_adults", "adults", "professionals", "general", "families"},
				"default":     "general",
			},
			"genre": map[string]interface{}{
				"type":        "string",
				"description": "Story genre",
				"enum":        []interface{}{"science_fiction", "slice_of_life", "adventure", "drama", "educational", "utopian", "mystery", "romance"},
				"default":     "science_fiction",
			},
			"story_length": map[string]interface{}{
				"type":        "string",
				"description": "Desired story length",
				"enum":        []interface{}{"short", "medium", "long"},
				"default":     "medium",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session ID for story tracking",
				"maxLength":   100,
			},
		},
		"required":    []string{"story_prompt"},
		"description": "Input schema for enhanced Storymaker agent",
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Story title",
				"minLength":   5,
				"maxLength":   150,
			},
			"story": map[string]interface{}{
				"type":        "string",
				"description": "The complete story narrative",
				"minLength":   300,
				"maxLength":   8000,
			},
			"themes": map[string]interface{}{
				"type":        "array",
				"description": "Key themes explored in the story",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 3,
					"maxLength": 80,
				},
			},
			"sustainability_features": map[string]interface{}{
				"type":        "array",
				"description": "Sustainability features highlighted in the story",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 10,
					"maxLength": 200,
				},
			},
			"moral": map[string]interface{}{
				"type":        "string",
				"description": "Key takeaway or moral of the story",
				"minLength":   20,
				"maxLength":   500,
			},
			"target_audience": map[string]interface{}{
				"type":        "string",
				"description": "Intended audience for the story",
				"enum":        []interface{}{"children", "young_adults", "adults", "professionals", "general", "families"},
			},
			"genre": map[string]interface{}{
				"type":        "string",
				"description": "Story genre",
				"enum":        []interface{}{"science_fiction", "slice_of_life", "adventure", "drama", "educational", "utopian", "mystery", "romance"},
			},
			"characters": map[string]interface{}{
				"type":        "array",
				"description": "Characters featured in the story",
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
							"maxLength": 300,
						},
						"motivation": map[string]interface{}{
							"type":      "string",
							"maxLength": 200,
						},
					},
					"required": []string{"name", "role"},
				},
			},
			"setting": map[string]interface{}{
				"type":        "object",
				"description": "Story setting details",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type": "string",
					},
					"time_period": map[string]interface{}{
						"type": "string",
					},
					"habitat_type": map[string]interface{}{
						"type": "string",
					},
					"environment": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"word_count": map[string]interface{}{
				"type":        "integer",
				"description": "Approximate word count of the story",
				"minimum":     50,
			},
			"reading_time": map[string]interface{}{
				"type":        "string",
				"description": "Estimated reading time",
			},
			"educational_value": map[string]interface{}{
				"type":        "string",
				"description": "Educational aspects of the story",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session ID for tracking",
			},
		},
		"required":    []string{"title", "story", "themes", "moral", "session_id"},
		"description": "Output schema for enhanced Storymaker agent",
	})

	definition.SetMetadata("description", "Creates engaging narratives about future sustainable habitat scenarios with educational value")
	definition.SetMetadata("tags", []string{"storytelling", "narrative", "futures", "sustainability", "education", "stateful"})

	return definition
}

// CreateAgent creates an enhanced Storymaker agent with story workflow
func (s *StorymakingDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := s.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build enhanced workflow graph with story creation processing
	graph := core.NewGraph("enhanced-storymaker-workflow")

	// Core workflow nodes
	graph.AddNode("create_story", "Create Story", s.createStoryNode)
	graph.SetStartNode("create_story")
	graph.AddEndNode("create_story")

	return baseAgent, nil
}

// createStoryNode handles the complete story creation process
func (s *StorymakingDefinition) createStoryNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

	input := inputData.(map[string]interface{})
	storyPrompt := input["story_prompt"].(string)

	sessionID, _ := input["session_id"].(string)
	if sessionID == "" {
		sessionID = fmt.Sprintf("story_%d", time.Now().Unix())
	}

	// Parse optional inputs
	targetAudience, _ := input["target_audience"].(string)
	if targetAudience == "" {
		targetAudience = "general"
	}

	genre, _ := input["genre"].(string)
	if genre == "" {
		genre = "science_fiction"
	}

	storyLength, _ := input["story_length"].(string)
	if storyLength == "" {
		storyLength = "medium"
	}

	// Initialize story context
	storyContext := &StoryContext{
		SessionID:       sessionID,
		StoriesCreated:  1,
		PreferredGenres: []string{},
		CharacterBank:   []StoryCharacter{},
		ThemeHistory:    []string{},
		LastCreation:    time.Now(),
		QualityScore:    0.0,
	}

	// Try to load existing story context
	if s.checkpointer != nil {
		if savedCheckpoint, err := s.checkpointer.Load(ctx, sessionID, sessionID); err == nil && savedCheckpoint != nil {
			if savedContextData, exists := savedCheckpoint.State.Get("story_context"); exists {
				if savedContext, ok := savedContextData.(*StoryContext); ok {
					storyContext = savedContext
					storyContext.StoriesCreated++
					storyContext.LastCreation = time.Now()
				}
			}
		}
	}

	// Parse setting information
	setting := s.parseStorySetting(input)

	// Parse character information
	characters := s.parseStoryCharacters(input, storyContext)

	// Create the story
	storyOutput := s.generateStory(storyPrompt, targetAudience, genre, storyLength, setting, characters)
	storyOutput.SessionID = sessionID
	storyOutput.ID = fmt.Sprintf("story_%d", time.Now().Unix())
	storyOutput.CreatedAt = time.Now()

	// Update story context
	s.updateStoryContext(storyContext, storyOutput)

	// Calculate quality metrics
	storyOutput.WordCount = s.calculateWordCount(storyOutput.Story)
	storyOutput.ReadingTime = s.calculateReadingTime(storyOutput.WordCount)
	storyOutput.EducationalValue = s.assessEducationalValue(storyOutput)

	// Save story context using checkpointer
	if s.checkpointer != nil {
		checkpoint := &persistence.Checkpoint{
			ID:        sessionID,
			ThreadID:  sessionID,
			State:     state,
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
			NodeID:    "story_creation",
			StepID:    storyContext.StoriesCreated,
		}
		state.Set("story_context", storyContext)
		s.checkpointer.Save(ctx, checkpoint)
	}

	// Structure final output
	output := map[string]interface{}{
		"title":                   storyOutput.Title,
		"story":                   storyOutput.Story,
		"themes":                  storyOutput.Themes,
		"sustainability_features": storyOutput.SustainabilityFeatures,
		"moral":                   storyOutput.Moral,
		"target_audience":         storyOutput.TargetAudience,
		"genre":                   storyOutput.Genre,
		"characters":              storyOutput.Characters,
		"setting":                 storyOutput.Setting,
		"word_count":              storyOutput.WordCount,
		"reading_time":            storyOutput.ReadingTime,
		"educational_value":       storyOutput.EducationalValue,
		"session_id":              storyOutput.SessionID,
	}

	state.Set("output", output)
	return state, nil
}

// Story creation helper methods

func (s *StorymakingDefinition) parseStorySetting(input map[string]interface{}) StorySetting {
	setting := StorySetting{
		Location:      "Eco-Community",
		TimePeriod:    "2035",
		HabitatType:   "community",
		Environment:   "mixed",
		SocialContext: "sustainable living community",
	}

	if settingData, ok := input["setting"].(map[string]interface{}); ok {
		if location, ok := settingData["location"].(string); ok && location != "" {
			setting.Location = location
		}
		if timePeriod, ok := settingData["time_period"].(string); ok && timePeriod != "" {
			setting.TimePeriod = timePeriod
		}
		if habitatType, ok := settingData["habitat_type"].(string); ok && habitatType != "" {
			setting.HabitatType = habitatType
		}
		if environment, ok := settingData["environment"].(string); ok && environment != "" {
			setting.Environment = environment
		}
	}

	return setting
}

func (s *StorymakingDefinition) parseStoryCharacters(input map[string]interface{}, context *StoryContext) []StoryCharacter {
	var characters []StoryCharacter

	if charactersData, ok := input["characters"].([]interface{}); ok {
		for _, charInterface := range charactersData {
			if charData, ok := charInterface.(map[string]interface{}); ok {
				character := StoryCharacter{}
				if name, ok := charData["name"].(string); ok {
					character.Name = name
				}
				if role, ok := charData["role"].(string); ok {
					character.Role = role
				}
				if background, ok := charData["background"].(string); ok {
					character.Background = background
				}
				if motivation, ok := charData["motivation"].(string); ok {
					character.Motivation = motivation
				}
				characters = append(characters, character)
			}
		}
	}

	// If no characters provided, create default characters
	if len(characters) == 0 {
		characters = s.createDefaultCharacters()
	}

	// Add to character bank for future stories
	context.CharacterBank = append(context.CharacterBank, characters...)

	return characters
}

func (s *StorymakingDefinition) createDefaultCharacters() []StoryCharacter {
	return []StoryCharacter{
		{
			Name:       "Maya",
			Role:       "Community Designer",
			Background: "Architect specializing in sustainable habitat design",
			Motivation: "Create beautiful, eco-friendly living spaces that bring people together",
		},
		{
			Name:       "Alex",
			Role:       "Sustainability Engineer",
			Background: "Environmental engineer focused on renewable energy systems",
			Motivation: "Develop innovative solutions for carbon-neutral living",
		},
		{
			Name:       "Luna",
			Role:       "Young Resident",
			Background: "Teenager who grew up in sustainable communities",
			Motivation: "Inspire others to embrace eco-friendly lifestyles",
		},
	}
}

func (s *StorymakingDefinition) generateStory(prompt, targetAudience, genre, length string, setting StorySetting, characters []StoryCharacter) StoryOutput {
	// Build comprehensive story generation prompt
	storyPrompt := s.buildStoryPrompt(prompt, setting, characters, targetAudience, genre, length)

	// Generate the story using LLM (simplified for now)
	storyText := s.generateStoryText(storyPrompt, length)

	// Extract themes from the story
	themes := s.extractStoryThemes(storyText, prompt)

	// Identify sustainability features
	sustainabilityFeatures := s.extractSustainabilityFeatures(storyText)

	// Generate moral/lesson
	moral := s.generateMoral(storyText, themes)

	// Generate title
	title := s.generateTitle(storyText, genre, themes)

	return StoryOutput{
		Title:                  title,
		Story:                  storyText,
		Themes:                 themes,
		SustainabilityFeatures: sustainabilityFeatures,
		Moral:                  moral,
		TargetAudience:         targetAudience,
		Genre:                  genre,
		Characters:             characters,
		Setting:                setting,
		Metadata:               make(map[string]interface{}),
	}
}

func (s *StorymakingDefinition) buildStoryPrompt(prompt string, setting StorySetting, characters []StoryCharacter, audience, genre, length string) string {
	characterDescriptions := ""
	for i, char := range characters {
		if i > 0 {
			characterDescriptions += ", "
		}
		characterDescriptions += fmt.Sprintf("%s (%s)", char.Name, char.Role)
	}

	lengthGuidelines := map[string]string{
		"short":  "300-600 mots",
		"medium": "600-1200 mots",
		"long":   "1200-2000 mots",
	}

	return fmt.Sprintf(`Créez une histoire captivante sur "%s"

Paramètres de l'histoire:
- Lieu: %s en %s
- Type d'habitat: %s dans un environnement %s
- Personnages principaux: %s
- Public cible: %s
- Genre: %s
- Longueur souhaitée: %s

Instructions spéciales:
- Intégrez des éléments de durabilité et d'habitat écologique
- Créez une narrative engageante avec des émotions authentiques
- Montrez les avantages pratiques et humains de la vie durable
- Incluez des détails techniques réalistes pour 2035
- Développez les personnages avec des motivations claires
- Terminez par un message inspirant et optimiste

L'histoire doit être à la fois divertissante et éducative, montrant comment la technologie et l'écologie peuvent créer des communautés harmonieuses.`,
		prompt,
		setting.Location,
		setting.TimePeriod,
		setting.HabitatType,
		setting.Environment,
		characterDescriptions,
		audience,
		genre,
		lengthGuidelines[length])
}

func (s *StorymakingDefinition) generateStoryText(prompt, length string) string {
	// Placeholder for LLM integration
	_ = []llm.Message{{Role: "user", Content: prompt}}

	// Generate story based on length
	switch length {
	case "short":
		return s.generateShortStory()
	case "long":
		return s.generateLongStory()
	default:
		return s.generateMediumStory()
	}
}

func (s *StorymakingDefinition) generateShortStory() string {
	return `🌱 L'Éveil de la Cité Verte

Maya se tenait sur la terrasse de sa maison en bambou, observant les premiers rayons du soleil alimenter les panneaux solaires intégrés au toit. En 2035, l'éco-quartier de Verdancia était devenu un modèle de vie durable.

"Regarde, Luna !" dit-elle à sa jeune voisine. "Nos jardins verticaux produisent déjà les légumes pour le petit-déjeuner."

Luna, maintenant étudiante en architecture écologique, sourit. "Maya, vous avez transformé notre façon de vivre. Chaque maison produit plus d'énergie qu'elle n'en consomme, et nos systèmes de récupération d'eau nous rendent complètement autonomes."

Alex, l'ingénieur en durabilité, les rejoignit avec son café dans une tasse en matériaux recyclés. "Ce matin, j'ai reçu une demande de consultation pour reproduire notre modèle en Asie. Verdancia inspire le monde entier."

Les trois amis regardèrent leur communauté s'éveiller : des enfants jouant dans les jardins partagés, des adultes se rendant au travail en véhicules électriques autonomes, et la nature qui prospérait en harmonie avec la technologie.

"Nous avons prouvé qu'il est possible de vivre mieux tout en respectant notre planète," murmura Maya. "Et ce n'est que le début."

Dans cette nouvelle ère, l'habitat durable n'était plus un rêve, mais une réalité joyeuse et prospère.`
}

func (s *StorymakingDefinition) generateMediumStory() string {
	return `🏡 La Maison qui Respirait avec la Terre

En 2035, dans l'éco-village de Terra Nova, Maya découvrait que sa nouvelle maison était bien plus qu'un simple abri. Construite avec des matériaux bio-sourcés et dotée d'une intelligence artificielle bienveillante, elle s'adaptait aux besoins de ses habitants et aux cycles naturels.

"Bonjour Maya," murmura la douce voix de Terra, l'IA de la maison. "L'air extérieur est particulièrement pur ce matin. Souhaitez-vous que j'ouvre les panneaux de ventilation naturelle?"

Maya sourit en préparant son thé avec l'eau purifiée par le système de phytoépuration intégré. "Oui, Terra. Et peux-tu ajuster l'éclairage pour optimiser ma séance de méditation?"

Les murs en terre crue et chanvre régulaient naturellement l'humidité, tandis que le toit végétalisé abritait des ruches connectées qui informaient Terra de la santé de l'écosystème local. Chaque élément de la maison contribuait à un cycle vertueux.

Alex, son voisin ingénieur, frappa à la porte. "Maya, nos maisons ont généré 15% d'énergie excédentaire ce mois-ci. Le réseau communautaire redistribue automatiquement cette énergie vers l'école et le centre médical."

Ensemble, ils se promenèrent dans le village où chaque habitation était unique, adaptée aux goûts de ses occupants mais partageant les mêmes principes durables. Les jardins-forêts nourrissaient la communauté, les ateliers de réparation prolongeaient la vie des objets, et les espaces partagés renforçaient les liens sociaux.

Luna, la jeune urbaniste, les rejoignit près du lac d'épuration naturelle. "J'ai terminé les plans pour l'extension du village. Nous pourrons accueillir 200 familles supplémentaires sans augmenter notre empreinte carbone."

"Comment est-ce possible?" demanda Maya, fascinée.

"En optimisant les synergies. Les eaux grises de chaque maison nourrissent les jardins communautaires, qui à leur tour purifient l'air et régulent le climat local. Les déchets organiques alimentent les digesteurs qui produisent le biogaz pour la cuisine. Et nos toits solaires partagent leur énergie via un micro-réseau intelligent."

Le soir venu, la communauté se rassemblait dans l'amphithéâtre naturel pour partager le repas préparé avec les produits locaux. Maya contemplait les étoiles, visible grâce à l'absence de pollution lumineuse.

"Nous avons créé plus qu'un habitat," réfléchit-elle. "Nous avons inventé une nouvelle façon de vivre en harmonie avec la nature, où la technologie amplifie notre humanité au lieu de la diminuer."

Terra Nova prouvait que l'avenir pouvait être à la fois high-tech et profondément humain.`
}

func (s *StorymakingDefinition) generateLongStory() string {
	return `🌍 L'Archipel des Rêves Durables

L'année 2035 avait apporté des changements extraordinaires. Sur l'archipel artificiel de Gaia, construit à partir de matériaux recyclés et d'algues biomimétiques, Maya contemplait l'océan depuis sa maison flottante. Cette communauté unique prouvait qu'il était possible de vivre sur l'eau tout en régénérant les écosystèmes marins.

"Maya, les coraux artificiels que nous avons plantés l'année dernière hébergent maintenant plus de cinquante espèces de poissons," annonça Alex en montrant les données de surveillance en temps réel sur son écran holographique. En tant qu'ingénieur en biomimétisme, il avait conçu des habitats qui non seulement ne nuisaient pas à l'océan, mais l'aidaient à guérir.

Leur communauté de 500 habitants vivait dans des maisons qui s'adaptaient aux marées et aux tempêtes, équipées de systèmes de dessalement alimentés par l'énergie des vagues et du vent. Chaque habitation était un écosystème vivant : les murs en algues purifiaient l'air, les sols en mycélium recyclaient les déchets organiques, et les toits cultivaient des jardins aériens.

Luna, maintenant docteure en écologie sociale, dirigeait le programme éducatif de l'archipel. "Nos visiteurs de cette semaine viennent de six continents différents. Ils veulent comprendre comment nous avons créé une société post-carbone qui améliore effectivement l'environnement."

Dans le laboratoire communautaire, les enfants apprenaient en fabriquant des matériaux de construction à partir de déchets plastiques et de coquillages. Les adultes développaient de nouvelles techniques d'aquaculture qui nourrissaient les familles tout en créant des récifs artificiels pour la biodiversité marine.

"Souviens-toi de nos débuts," dit Maya à Alex en marchant sur les pontons de bambou qui reliaient les îlots. "Nous pensions qu'il suffisait de construire des maisons écologiques. Maintenant, nous réalisons que nous créons des organismes vivants qui évoluent avec nous."

Chaque maison était équipée d'IA symbiotiques qui apprenaient les habitudes de leurs habitants et optimisaient automatiquement la consommation d'énergie, la qualité de l'air et même l'éclairage circadien pour améliorer la santé. Les jardins flottants étaient cultivés par des robots jardiniers qui pollinisaient également les coraux artificiels.

Le centre communautaire, construit en forme de spirale pour maximiser la circulation naturelle de l'air, accueillait ce soir-là le conseil hebdomadaire. Les décisions se prenaient par consensus, assisté par une IA qui modélisait l'impact environnemental et social de chaque proposition.

"Nous devons voter sur l'invitation du gouvernement indonésien," annonça le facilitateur. "Ils nous demandent d'essaimer notre modèle sur dix sites dans l'archipel."

Luna prit la parole : "C'est exactement notre mission. Mais nous devons nous assurer que chaque nouvelle communauté s'adapte à son écosystème local unique. Gaia n'est pas un modèle à copier, c'est un principe à réinventer."

Après le vote unanime en faveur de l'expansion, la communauté célébrait sur la plage bioluminescente, où des micro-organismes génétiquement modifiés créaient un spectacle de lumière naturelle sans électricité.

Maya regardait les enfants jouer dans l'eau chaude des lagons artificiels, où poissons tropicaux et dauphins coexistaient pacifiquement avec les humains. "Alex, tu te rappelles quand les gens disaient que nous étions utopistes?"

"Maintenant ils disent que nous sommes la nouvelle normalité," sourit Alex. "Trois cent cinquante communautés similaires existent déjà sur les sept continents."

En s'endormant dans sa chambre aux murs qui respiraient et se régénéraient, Maya écoutait le doux bruissement des vagues contre les fondations vivantes de sa maison. Demain, une nouvelle équipe d'étudiants arriverait pour apprendre à créer des habitats qui nourrissent la planète.

L'archipel de Gaia n'était plus un rêve, mais la preuve vivante qu'un autre monde était non seulement possible, mais déjà en train de naître.`
}

func (s *StorymakingDefinition) extractStoryThemes(story, prompt string) []string {
	themes := []string{}

	storyLower := strings.ToLower(story)
	promptLower := strings.ToLower(prompt)

	themeKeywords := map[string]string{
		"sustainability": "Durabilité et Écologie",
		"community":      "Communauté et Coopération",
		"technology":     "Technologie et Innovation",
		"harmony":        "Harmonie avec la Nature",
		"future":         "Vision du Futur",
		"education":      "Éducation et Apprentissage",
		"energy":         "Énergie Renouvelable",
		"housing":        "Habitat Intelligent",
		"food":           "Alimentation Durable",
		"water":          "Gestion de l'Eau",
		"biodiversity":   "Biodiversité et Conservation",
		"innovation":     "Innovation Écologique",
	}

	for keyword, theme := range themeKeywords {
		if strings.Contains(storyLower, keyword) || strings.Contains(promptLower, keyword) {
			themes = append(themes, theme)
		}
	}

	// Add base themes
	themes = append(themes, "Habitat Durable 2035")

	return s.deduplicateThemes(themes)
}

func (s *StorymakingDefinition) extractSustainabilityFeatures(story string) []string {
	features := []string{}
	storyLower := strings.ToLower(story)

	sustainabilityFeatures := map[string]string{
		"solar":     "Panneaux solaires intégrés",
		"wind":      "Énergie éolienne",
		"water":     "Systèmes de récupération d'eau",
		"garden":    "Jardins verticaux et permaculture",
		"recycl":    "Matériaux recyclés et circulaires",
		"bamboo":    "Construction en bambou",
		"biomas":    "Bioénergie et biomasse",
		"carbon":    "Neutralité carbone",
		"ecosystem": "Préservation des écosystèmes",
		"compost":   "Compostage et gestion des déchets",
		"green":     "Technologies vertes",
		"renewable": "Énergies renouvelables",
		"electric":  "Véhicules électriques",
		"natural":   "Ventilation naturelle",
		"organic":   "Agriculture biologique",
	}

	for keyword, feature := range sustainabilityFeatures {
		if strings.Contains(storyLower, keyword) {
			features = append(features, feature)
		}
	}

	if len(features) == 0 {
		features = append(features, "Conception écologique intégrée")
	}

	return features
}

func (s *StorymakingDefinition) generateMoral(story string, themes []string) string {
	morals := []string{
		"L'avenir durable commence par les choix que nous faisons aujourd'hui pour nos habitats.",
		"La technologie et la nature peuvent créer ensemble des communautés harmonieuses et prospères.",
		"Vivre durablement enrichit notre qualité de vie et protège notre planète pour les générations futures.",
		"L'innovation écologique transforme nos maisons en écosystèmes vivants qui nous nourrissent.",
		"Les communautés durables prouvent qu'un autre mode de vie est possible et accessible.",
	}

	// Select moral based on story content
	if strings.Contains(strings.ToLower(story), "community") {
		return morals[4]
	} else if strings.Contains(strings.ToLower(story), "technology") {
		return morals[1]
	} else if strings.Contains(strings.ToLower(story), "ecosystem") {
		return morals[3]
	}

	return morals[0] // Default moral
}

func (s *StorymakingDefinition) generateTitle(story, genre string, themes []string) string {
	titles := map[string][]string{
		"science_fiction": {
			"L'Archipel des Rêves Durables",
			"Terra Nova: L'Éveil Écologique",
			"Les Jardins Flottants de 2035",
			"Gaia: L'Habitat qui Respirait",
		},
		"slice_of_life": {
			"La Maison qui Respirait avec la Terre",
			"Un Matin dans l'Éco-Village",
			"Les Voix de la Communauté Verte",
			"Habiter l'Avenir Durable",
		},
		"educational": {
			"Leçons de Vie Durable",
			"L'École de l'Habitat Écologique",
			"Apprendre à Vivre en 2035",
			"Les Secrets de l'Éco-Construction",
		},
		"adventure": {
			"La Quête de l'Habitat Parfait",
			"Aventures dans la Cité Verte",
			"L'Exploration de Terra Futura",
			"Mission Durabilité 2035",
		},
	}

	if genreTitles, exists := titles[genre]; exists {
		// Select title based on story content
		for i, title := range genreTitles {
			if i == 0 { // Return first title for now
				return title
			}
		}
	}

	return "Récits d'Habitats Durables 2035" // Default title
}

func (s *StorymakingDefinition) updateStoryContext(context *StoryContext, story StoryOutput) {
	// Update preferred genres
	if !s.containsString(context.PreferredGenres, story.Genre) {
		context.PreferredGenres = append(context.PreferredGenres, story.Genre)
	}

	// Update theme history
	for _, theme := range story.Themes {
		if !s.containsString(context.ThemeHistory, theme) {
			context.ThemeHistory = append(context.ThemeHistory, theme)
		}
	}

	// Calculate quality score
	context.QualityScore = s.calculateStoryQuality(story)
}

func (s *StorymakingDefinition) calculateWordCount(story string) int {
	words := strings.Fields(story)
	return len(words)
}

func (s *StorymakingDefinition) calculateReadingTime(wordCount int) string {
	// Average reading speed: 200 words per minute
	minutes := wordCount / 200
	if minutes < 1 {
		return "Moins d'1 minute"
	} else if minutes == 1 {
		return "1 minute"
	} else {
		return fmt.Sprintf("%d minutes", minutes)
	}
}

func (s *StorymakingDefinition) assessEducationalValue(story StoryOutput) string {
	educationalAspects := []string{}

	if len(story.SustainabilityFeatures) >= 3 {
		educationalAspects = append(educationalAspects, "Concepts de durabilité")
	}

	if len(story.Characters) >= 2 {
		educationalAspects = append(educationalAspects, "Collaboration communautaire")
	}

	for _, theme := range story.Themes {
		if theme == "Technologie et Innovation" {
			educationalAspects = append(educationalAspects, "Innovation technologique")
			break
		}
	}

	if len(educationalAspects) == 0 {
		return "Sensibilisation écologique de base"
	}

	return strings.Join(educationalAspects, ", ")
}

func (s *StorymakingDefinition) calculateStoryQuality(story StoryOutput) float64 {
	score := 0.0

	// Word count score
	if story.WordCount >= 300 && story.WordCount <= 2000 {
		score += 0.2
	}

	// Theme diversity
	score += float64(len(story.Themes)) * 0.1

	// Sustainability features
	score += float64(len(story.SustainabilityFeatures)) * 0.1

	// Character development
	score += float64(len(story.Characters)) * 0.1

	// Has meaningful moral
	if len(story.Moral) >= 20 {
		score += 0.2
	}

	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// Utility methods

func (s *StorymakingDefinition) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *StorymakingDefinition) deduplicateThemes(themes []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, theme := range themes {
		if !seen[theme] {
			seen[theme] = true
			result = append(result, theme)
		}
	}

	return result
}

// GetStorymakerConfig returns the configuration for backward compatibility
func GetStorymakerConfig() *agent.AgentConfig {
	// For backward compatibility, create a temporary instance
	temp := NewStorymakerDefinition()
	return temp.GetConfig()
}
