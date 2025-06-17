package weather

// WeatherEmoji provides an emoji based on temperature.
type WeatherEmoji struct{}

// Emoji returns an emoji string based on the given temperature.
func (*WeatherEmoji) Emoji(temp float64) string {
    switch {
    case temp >= 35:
        return "🥵" // very hot
    case temp >= 25:
        return "😎" // warm
    case temp >= 15:
        return "🙂" // mild
    case temp >= 5:
        return "🧥" // cool
    default:
        return "🥶" // cold
    }
}
