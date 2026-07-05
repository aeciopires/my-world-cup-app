package handlers

import "github.com/aeciopires/my-world-cup-app/internal/models"

const fifaTournamentBaseURL = "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/"

// fifaTeamSlugs maps each team's name (models.Team.Name, as carried by
// openfootball) to its fifa.com team page slug. FIFA's slugs don't always
// match the openfootball name (accents, official renames such as "Turkey"
// -> "turkiye" or "South Korea" -> "korea-republic"), so this is a manual
// lookup rather than a derived slugify.
var fifaTeamSlugs = map[string]string{
	"Mexico":               "mexico",
	"South Africa":         "south-africa",
	"South Korea":          "korea-republic",
	"Czech Republic":       "czechia",
	"Canada":               "canada",
	"Bosnia & Herzegovina": "bosnia-herzegovina",
	"Qatar":                "qatar",
	"Switzerland":          "switzerland",
	"Brazil":               "brazil",
	"Morocco":              "morocco",
	"Haiti":                "haiti",
	"Scotland":             "scotland",
	"USA":                  "usa",
	"Paraguay":             "paraguay",
	"Australia":            "australia",
	"Turkey":               "turkiye",
	"Germany":              "germany",
	"Curaçao":              "curacao",
	"Ivory Coast":          "cote-d-ivoire",
	"Ecuador":              "ecuador",
	"Netherlands":          "netherlands",
	"Japan":                "japan",
	"Sweden":               "sweden",
	"Tunisia":              "tunisia",
	"Belgium":              "belgium",
	"Egypt":                "egypt",
	"Iran":                 "ir-iran",
	"New Zealand":          "new-zealand",
	"Spain":                "spain",
	"Cape Verde":           "cabo-verde",
	"Saudi Arabia":         "saudi-arabia",
	"Uruguay":              "uruguay",
	"France":               "france",
	"Senegal":              "senegal",
	"Iraq":                 "iraq",
	"Norway":               "norway",
	"Argentina":            "argentina",
	"Algeria":              "algeria",
	"Austria":              "austria",
	"Jordan":               "jordan",
	"Portugal":             "portugal",
	"DR Congo":             "congo-dr",
	"Uzbekistan":           "uzbekistan",
	"Colombia":             "colombia",
	"England":              "england",
	"Croatia":              "croatia",
	"Ghana":                "ghana",
	"Panama":               "panama",
}

// fifaHostCitySlugs maps each stadium's host city (models.Stadium.City) to
// the fifa.com path segment shared by both the stadium page
// (stadiums/<slug>) and the host-city page (<country>/<slug>). FIFA
// occasionally renames a stadium for marketing reasons (e.g. Estadio Azteca),
// but the host-city slug stays stable, so lookups key off the city rather
// than the stadium's own name.
var fifaHostCitySlugs = map[string]string{
	"Vancouver":                             "vancouver",
	"Seattle":                               "seattle",
	"San Francisco Bay Area (Santa Clara)":  "san-francisco-bay-area",
	"Los Angeles (Inglewood)":               "los-angeles",
	"Guadalajara (Zapopan)":                 "guadalajara",
	"Mexico City":                           "mexico-city",
	"Monterrey (Guadalupe)":                 "monterrey",
	"Houston":                               "houston",
	"Dallas (Arlington)":                    "dallas",
	"Kansas City":                           "kansas-city",
	"Atlanta":                               "atlanta",
	"Miami (Miami Gardens)":                 "miami",
	"Toronto":                               "toronto",
	"Boston (Foxborough)":                   "boston",
	"Philadelphia":                          "philadelphia",
	"New York/New Jersey (East Rutherford)": "new-york-new-jersey",
}

// fifaHostCountrySlugs maps a stadium's country code (models.Stadium.Country,
// e.g. "ca"/"mx"/"us") to the fifa.com host country path segment used in
// host-city page URLs.
var fifaHostCountrySlugs = map[string]string{
	"ca": "canada",
	"mx": "mexico",
	"us": "usa",
}

// teamLinks maps each team's name to its official fifa.com team news page.
func teamLinks(t models.Tournament) map[string]string {
	links := make(map[string]string, len(t.Teams))
	for _, team := range t.Teams {
		if slug, ok := fifaTeamSlugs[team.Name]; ok {
			links[team.Name] = fifaTournamentBaseURL + "teams/" + slug + "/team-news"
		}
	}
	return links
}

// stadiumLinks maps each stadium's name to its official fifa.com stadium page.
func stadiumLinks(t models.Tournament) map[string]string {
	links := make(map[string]string, len(t.Stadiums))
	for _, s := range t.Stadiums {
		if slug, ok := fifaHostCitySlugs[s.City]; ok {
			links[s.Name] = fifaTournamentBaseURL + "stadiums/" + slug
		}
	}
	return links
}

// cityLinks maps each host city's name to its official fifa.com host-city page.
func cityLinks(t models.Tournament) map[string]string {
	links := make(map[string]string, len(t.Stadiums))
	for _, s := range t.Stadiums {
		slug, ok := fifaHostCitySlugs[s.City]
		country, ok2 := fifaHostCountrySlugs[s.Country]
		if ok && ok2 {
			links[s.City] = fifaTournamentBaseURL + country + "/" + slug
		}
	}
	return links
}
