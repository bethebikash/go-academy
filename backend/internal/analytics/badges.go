package analytics

// AvailableBadges defines all platform achievements.
var AvailableBadges = []BadgeDefinition{
	{
		Key:         BadgeFirstStep,
		Name:        "First Step",
		Description: "Complete your first Go curriculum lesson.",
		Category:    "Curriculum",
		Icon:        "Sparkles",
	},
	{
		Key:         BadgeBugHunter,
		Name:        "Bug Hunter",
		Description: "Solve and pass all test cases on a Go coding challenge.",
		Category:    "Challenges",
		Icon:        "CheckCircle2",
	},
	{
		Key:         BadgeScholar,
		Name:        "Go Scholar",
		Description: "Take 5 or more study notes on lessons or challenges.",
		Category:    "Study",
		Icon:        "BookOpen",
	},
	{
		Key:         BadgeStreak7,
		Name:        "Week of Code",
		Description: "Maintain an active 7-day learning streak.",
		Category:    "Dedication",
		Icon:        "Flame",
	},
	{
		Key:         BadgePathMaster,
		Name:        "Path Master",
		Description: "Complete 100% of an entire Go learning path.",
		Category:    "Mastery",
		Icon:        "Award",
	},
}
