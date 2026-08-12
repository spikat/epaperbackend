package weather

func IconName(code int) string {
	switch {
	case code == 0:
		return "clear"
	case code == 1:
		return "mainly_clear"
	case code == 2:
		return "partly_cloudy"
	case code == 3:
		return "cloudy"
	case code >= 45 && code <= 48:
		return "fog"
	case code >= 51 && code <= 57:
		return "drizzle"
	case code >= 61 && code <= 67:
		return "rain"
	case code >= 71 && code <= 77:
		return "snow"
	case code >= 80 && code <= 82:
		return "showers"
	case code >= 95:
		return "thunderstorm"
	default:
		return "cloudy"
	}
}

var iconSVGs = map[string]string{
	"clear": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><circle cx="32" cy="32" r="14" fill="#000" stroke="#000" stroke-width="2"/><g stroke="#000" stroke-width="3"><line x1="32" y1="4" x2="32" y2="12"/><line x1="32" y1="52" x2="32" y2="60"/><line x1="4" y1="32" x2="12" y2="32"/><line x1="52" y1="32" x2="60" y2="32"/><line x1="12" y1="12" x2="18" y2="18"/><line x1="46" y1="46" x2="52" y2="52"/><line x1="12" y1="52" x2="18" y2="46"/><line x1="46" y1="18" x2="52" y2="12"/></g></svg>`,
	"mainly_clear": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><circle cx="24" cy="24" r="10" fill="#000"/><path d="M18 40h28a10 10 0 0 0 0-20 12 12 0 0 0-23.5 3.5A8 8 0 0 0 18 40z" fill="#fff" stroke="#000" stroke-width="2"/></svg>`,
	"partly_cloudy": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><circle cx="22" cy="22" r="9" fill="#000"/><path d="M16 38h30a11 11 0 0 0 .5-22 13 13 0 0 0-25 4A9 9 0 0 0 16 38z" fill="#fff" stroke="#000" stroke-width="2"/></svg>`,
	"cloudy": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path d="M16 42h32a12 12 0 0 0 1-24 14 14 0 0 0-27.5 4A10 10 0 0 0 16 42z" fill="#fff" stroke="#000" stroke-width="2"/></svg>`,
	"fog": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path d="M12 24h40M8 34h48M14 44h36" stroke="#000" stroke-width="3" stroke-linecap="round"/></svg>`,
	"drizzle": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path d="M14 28h36a10 10 0 0 0 0-20 12 12 0 0 0-23.5 3.5A8 8 0 0 0 14 28z" fill="#fff" stroke="#000" stroke-width="2"/><g stroke="#000" stroke-width="2"><line x1="22" y1="36" x2="18" y2="44"/><line x1="32" y1="36" x2="28" y2="44"/><line x1="42" y1="36" x2="38" y2="44"/></g></svg>`,
	"rain": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path d="M14 26h36a10 10 0 0 0 0-20 12 12 0 0 0-23.5 3.5A8 8 0 0 0 14 26z" fill="#fff" stroke="#000" stroke-width="2"/><g stroke="#000" stroke-width="3"><line x1="20" y1="34" x2="14" y2="46"/><line x1="32" y1="34" x2="26" y2="46"/><line x1="44" y1="34" x2="38" y2="46"/></g></svg>`,
	"snow": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path d="M14 26h36a10 10 0 0 0 0-20 12 12 0 0 0-23.5 3.5A8 8 0 0 0 14 26z" fill="#fff" stroke="#000" stroke-width="2"/><g fill="#000"><circle cx="22" cy="40" r="2"/><circle cx="32" cy="44" r="2"/><circle cx="42" cy="40" r="2"/></g></svg>`,
	"showers": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path d="M12 24h40a11 11 0 0 0 0-22 13 13 0 0 0-25.5 4A9 9 0 0 0 12 24z" fill="#fff" stroke="#000" stroke-width="2"/><g stroke="#000" stroke-width="2"><line x1="24" y1="32" x2="20" y2="40"/><line x1="34" y1="32" x2="30" y2="40"/><line x1="44" y1="32" x2="40" y2="40"/></g></svg>`,
	"thunderstorm": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><path d="M14 24h36a10 10 0 0 0 0-20 12 12 0 0 0-23.5 3.5A8 8 0 0 0 14 24z" fill="#fff" stroke="#000" stroke-width="2"/><polygon points="34,30 26,42 32,42 28,54 40,38 34,38" fill="#000"/></svg>`,
}

func IconSVG(name string) (string, bool) {
	svg, ok := iconSVGs[name]
	return svg, ok
}
