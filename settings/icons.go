package settings

import (
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// settingsIcon parses an SVG once, on first use, as the icon package does,
// so the settings window does not root its category icons at start.
type settingsIcon struct {
	once sync.Once
	svg  *ui.SVG
}

func (i *settingsIcon) get(data string) *ui.SVG {
	i.once.Do(func() { i.svg = ui.MustParseSVG([]byte(data)) })
	return i.svg
}

// settingsCategoryIcon is the icon otty draws beside each settings category.
// The paths are otty's own, taken from its settings UI. The animation
// classes and the database icon's SMIL morph are dropped: mygo parses SVG
// statically, so they would never play, and the breathing is drawn with a
// mygo Keyframes pose on a click of the Data row instead (see
// breatheData). Profiles, which otty's settings list no longer shows, keeps
// a lucide user card in its place.
func settingsCategoryIcon(id string) *ui.SVG {
	switch id {
	case "general":
		return settingsCategoryIcon_general.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><g fill="currentColor" opacity=".05"><path d="M10.5 3.866a3 3 0 0 1 3 0l4.794 2.768a3 3 0 0 1 1.5 2.598v5.536a3 3 0 0 1-1.5 2.598L13.5 20.134a3 3 0 0 1-3 0l-4.794-2.768a3 3 0 0 1-1.5-2.598V9.232a3 3 0 0 1 1.5-2.598z"/><path d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0"/></g><path stroke="currentColor" stroke-width="2" d="M10.5 3.866a3 3 0 0 1 3 0l4.794 2.768a3 3 0 0 1 1.5 2.598v5.536a3 3 0 0 1-1.5 2.598L13.5 20.134a3 3 0 0 1-3 0l-4.794-2.768a3 3 0 0 1-1.5-2.598V9.232a3 3 0 0 1 1.5-2.598z"/><path stroke="currentColor" stroke-width="2" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/></svg>`)
	case "shell":
		return settingsCategoryIcon_shell.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M4 5a1 1 0 0 1 1-1h14a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1z" opacity=".05"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 20h14a1 1 0 0 0 1-1V5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v14a1 1 0 0 0 1 1"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.343 9.172 11.172 12l-2.829 2.828"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 15h2"/></svg>`)
	case "remote":
		return settingsCategoryIcon_remote.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M16 13H8a2 2 0 0 0-2 2v7h12v-7a2 2 0 0 0-2-2" opacity=".05"/><path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M18 21v-6a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v6M11 17h2"/><path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M14.285 9.717A3.982 3.982 0 0 0 12 9c-.85 0-1.638.265-2.285.717"/><path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M16.426 7.577A6.971 6.971 0 0 0 12 6c-1.679 0-3.219.59-4.425 1.576"/><path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M5.442 5.45A9.961 9.961 0 0 1 12 3a9.96 9.96 0 0 1 6.553 2.446"/></svg>`)
	case "controls":
		return settingsCategoryIcon_controls.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="m7.69 3.28-.847 13.334 3.521-3.042 3.152 7.45 2.705-1.152-3.018-7.212 4.732-.292z" opacity=".05"/><path stroke="currentColor" stroke-width="2" d="M7.844 3.416a.1.1 0 0 0-.166.069l-.82 12.891a.1.1 0 0 0 .165.082l3.237-2.796a.1.1 0 0 1 .158.036l2.903 6.863a.5.5 0 0 0 .656.265l1.787-.761a.5.5 0 0 0 .265-.653l-2.772-6.624a.1.1 0 0 1 .086-.139l4.351-.268a.1.1 0 0 0 .06-.175z"/></svg>`)
	case "editor":
		return settingsCategoryIcon_editor.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M5 4a1 1 0 0 1 1-1h7.586a1 1 0 0 1 .707.293l4.414 4.414a1 1 0 0 1 .293.707V20a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1z" opacity=".05"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 3v5.5a.5.5 0 0 0 .5.5H19m-5.414-6H6a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1V8.414a1 1 0 0 0-.293-.707l-4.414-4.414A1 1 0 0 0 13.586 3"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h1"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 16h6"/></svg>`)
	case "integrations", "agents":
		return settingsCategoryIcon_integrations.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M4 10a4 4 0 0 1 4-4h8a4 4 0 0 1 4 4v6a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4z" opacity=".05"/><path stroke="currentColor" stroke-linecap="round" stroke-width="2" d="M9 12v2m6-2v2m-3-9v2m0-2a1 1 0 1 0 0-2 1 1 0 0 0 0 2Zm-7 6h-.5a1.5 1.5 0 0 0 0 3H5zm14 0h.5a1.5 1.5 0 0 1 0 3H19zM8 19h8a3 3 0 0 0 3-3v-6a3 3 0 0 0-3-3H8a3 3 0 0 0-3 3v6a3 3 0 0 0 3 3Z"/></svg>`)
	case "view":
		return settingsCategoryIcon_view.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><g><g><path fill="currentColor" opacity=".05" d="M10 5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M10 5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/></g><g><path fill="currentColor" opacity=".05" d="M20 5a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M20 5a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/></g><g><path fill="currentColor" opacity=".05" d="M10 15a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M10 15a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/></g><g><path fill="currentColor" opacity=".05" d="M20 15a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M20 15a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1z"/></g></g></svg>`)
	case "appearance":
		return settingsCategoryIcon_appearance.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M17.902 15.484c1.322.22 2.682-.458 2.936-1.773a9 9 0 1 0-8.469 7.282c1.292-.053 1.891-1.472 1.313-2.63a2.115 2.115 0 0 1 .396-2.44l.089-.09a2.29 2.29 0 0 1 1.995-.64z" opacity=".05"/><path fill="currentColor" d="M8 12.5a.5.5 0 1 1-1 0 .5.5 0 0 1 1 0m2-4a.5.5 0 1 1-1 0 .5.5 0 0 1 1 0m5 0a.5.5 0 1 1-1 0 .5.5 0 0 1 1 0"/><path stroke="currentColor" stroke-width="2" d="M17.902 15.484c1.322.22 2.682-.458 2.936-1.773a9 9 0 1 0-8.469 7.282c1.292-.053 1.891-1.472 1.313-2.63a2.115 2.115 0 0 1 .396-2.44l.089-.09a2.29 2.29 0 0 1 1.995-.64z"/><path stroke="currentColor" stroke-width="2" d="M8 12.5a.5.5 0 1 1-1 0 .5.5 0 0 1 1 0Zm2-4a.5.5 0 1 1-1 0 .5.5 0 0 1 1 0Zm5 0a.5.5 0 1 1-1 0 .5.5 0 0 1 1 0Z"/></svg>`)
	case "recipes":
		return settingsCategoryIcon_recipes.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M3 6.5c0-.315.148-.61.417-.774C4.129 5.293 5.683 4.5 7.5 4.5 10 4.5 12 6 12 6v14s-2-1.5-4.5-1.5c-1.508 0-2.834.546-3.649.979C3.488 19.672 3 19.412 3 19zm18 0a.893.893 0 0 0-.417-.774C19.871 5.293 18.318 4.5 16.5 4.5 14 4.5 12 6 12 6v14s2-1.5 4.5-1.5c1.508 0 2.834.546 3.648.979.364.193.852-.067.852-.479z" opacity=".05"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M12 6v14m0-14s-2-1.5-4.5-1.5c-1.817 0-3.37.793-4.083 1.226A.893.893 0 0 0 3 6.5V19c0 .412.488.672.851.479.815-.433 2.141-.979 3.649-.979C10 18.5 12 20 12 20m0-14s2-1.5 4.5-1.5c1.817 0 3.37.793 4.083 1.226.27.163.417.46.417.774V19c0 .412-.488.672-.852.479-.814-.433-2.14-.979-3.648-.979C14 18.5 12 20 12 20"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M3 6.5c0-.315.148-.61.417-.774C4.129 5.293 5.683 4.5 7.5 4.5 10 4.5 12 6 12 6v14s-2-1.5-4.5-1.5c-1.508 0-2.834.546-3.649.979C3.488 19.672 3 19.412 3 19z"/></svg>`)
	case "keybindings", "key_bindings":
		return settingsCategoryIcon_keybindings.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M3 6a1 1 0 0 1 1-1h16a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1z" opacity=".05"/><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 9h1m3.5 0h1M16 9h1M7 12h1m3.5 0h1m3.5 0h1M7 15h10M4 19h16a1 1 0 0 0 1-1V6a1 1 0 0 0-1-1H4a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1"/></svg>`)
	case "data":
		return settingsCategoryIcon_data.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><g><path fill="currentColor" opacity=".05" d="M11.25 3.433a1.5 1.5 0 0 1 1.5 0l3.069 1.772 3.225 1.862a1.5 1.5 0 0 1 .75 1.299v7.268a1.5 1.5 0 0 1-.75 1.299l-6.294 3.634a1.5 1.5 0 0 1-1.5 0l-6.294-3.634a1.5 1.5 0 0 1-.75-1.299V8.366a1.5 1.5 0 0 1 .75-1.299z"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M11.25 3.433a1.5 1.5 0 0 1 1.5 0l3.069 1.772 3.225 1.862a1.5 1.5 0 0 1 .75 1.299v7.268a1.5 1.5 0 0 1-.75 1.299l-6.294 3.634a1.5 1.5 0 0 1-1.5 0l-6.294-3.634a1.5 1.5 0 0 1-.75-1.299V8.366a1.5 1.5 0 0 1 .75-1.299zM12 20.768V11.768M19.593 7.616 12.75 11.567a1.5 1.5 0 0 1-1.5 0L4.407 7.616"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M4.956 7.067C4.956 7.067 11.25 3.433 11.25 3.433C11.714 3.165 12.286 3.165 12.75 3.433C12.75 3.433 15.819 5.204 15.819 5.204C15.819 5.204 8.018 9.701 8.018 9.701C8.018 9.701 4.407 7.616 4.407 7.616C4.536 7.392 4.724 7.201 4.956 7.067Z"><animate attributeName="d" begin="indefinite" dur="0.983s" calcMode="spline" keyTimes="0;0.339;0.678;1" keySplines="0.05 0 0 1;0.05 0 0.95 1;1 0 0.95 1" values="M4.956 7.067C4.956 7.067 11.25 3.433 11.25 3.433C11.714 3.165 12.286 3.165 12.75 3.433C12.75 3.433 15.819 5.204 15.819 5.204C15.819 5.204 8.018 9.701 8.018 9.701C8.018 9.701 4.407 7.616 4.407 7.616C4.536 7.392 4.724 7.201 4.956 7.067Z;M4.956 7.067C4.956 7.067 11.25 3.433 11.25 3.433C11.58 3.243 12 2.957 12 2.576C12 2.576 12 -0.796 12 -0.796C12 -0.796 4.21 3.701 4.21 3.701C4.21 3.701 4.21 7.616 4.21 7.616C4.5 7.5 4.724 7.201 4.956 7.067Z;M4.956 7.067C4.956 7.067 11.25 3.433 11.25 3.433C11.58 3.243 12 2.957 12 2.576C12 2.576 12 -0.796 12 -0.796C12 -0.796 4.21 3.701 4.21 3.701C4.21 3.701 4.21 7.616 4.21 7.616C4.5 7.5 4.724 7.201 4.956 7.067Z;M4.956 7.067C4.956 7.067 11.25 3.433 11.25 3.433C11.714 3.165 12.286 3.165 12.75 3.433C12.75 3.433 15.819 5.204 15.819 5.204C15.819 5.204 8.018 9.701 8.018 9.701C8.018 9.701 4.407 7.616 4.407 7.616C4.536 7.392 4.724 7.201 4.956 7.067Z"/></path><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M15.819 5.204C15.819 5.204 8.018 9.701 8.018 9.701C8.018 9.701 11.25 11.567 11.25 11.567C11.714 11.835 12.286 11.835 12.75 11.567C12.75 11.567 19.593 7.616 19.593 7.616C19.464 7.392 19.276 7.201 19.044 7.067C19.044 7.067 15.819 5.204 15.819 5.204Z"><animate attributeName="d" begin="indefinite" dur="0.983s" calcMode="spline" keyTimes="0;0.339;0.678;1" keySplines="0.05 0 0 1;0.05 0 0.95 1;1 0 0.95 1" values="M15.819 5.204C15.819 5.204 8.018 9.701 8.018 9.701C8.018 9.701 11.25 11.567 11.25 11.567C11.714 11.835 12.286 11.835 12.75 11.567C12.75 11.567 19.593 7.616 19.593 7.616C19.464 7.392 19.276 7.201 19.044 7.067C19.044 7.067 15.819 5.204 15.819 5.204Z;M19.79 3.204C19.79 3.204 12 7.701 12 7.701C12 7.701 12 11.567 12 11.567C12.464 11.835 12.286 11.835 12.75 11.567C12.75 11.567 19.5 7.701 19.5 7.701C19.71 7.585 19.79 7.266 19.79 7.067C19.79 7.067 19.79 3.204 19.79 3.204Z;M19.79 3.204C19.79 3.204 12 7.701 12 7.701C12 7.701 12 11.567 12 11.567C12.464 11.835 12.286 11.835 12.75 11.567C12.75 11.567 19.5 7.701 19.5 7.701C19.71 7.585 19.79 7.266 19.79 7.067C19.79 7.067 19.79 3.204 19.79 3.204Z;M15.819 5.204C15.819 5.204 8.018 9.701 8.018 9.701C8.018 9.701 11.25 11.567 11.25 11.567C11.714 11.835 12.286 11.835 12.75 11.567C12.75 11.567 19.593 7.616 19.593 7.616C19.464 7.392 19.276 7.201 19.044 7.067C19.044 7.067 15.819 5.204 15.819 5.204Z"/></path></g></svg>`)
	case "advanced":
		return settingsCategoryIcon_advanced.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><path fill="currentColor" d="M13.646 5.14a6.002 6.002 0 0 1 1.267 6.62l4.928 4.17a2.76 2.76 0 1 1-3.89 3.89l-4.17-4.927a6.002 6.002 0 0 1-7.8-8.083l3.832 4.162 2.652-.527.53-2.655L6.83 3.96a6.003 6.003 0 0 1 6.816 1.18" opacity=".05"/><path stroke="currentColor" stroke-linejoin="round" stroke-width="2" d="M13.646 5.14a6.002 6.002 0 0 1 1.267 6.62l4.928 4.17a2.76 2.76 0 1 1-3.89 3.89l-4.17-4.927a6.002 6.002 0 0 1-7.8-8.083l3.832 4.162 2.652-.527.53-2.655L6.83 3.96a6.003 6.003 0 0 1 6.816 1.18Z"/></svg>`)
	case "account":
		return settingsCategoryIcon_account.get(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"><g><path fill="currentColor" opacity=".05" d="M20 18.5c0 1.933-3.582 2.5-8 2.5s-8-.567-8-2.5S7.582 14 12 14s8 2.567 8 4.5"/><path stroke="currentColor" stroke-width="2" d="M20 18.5c0 1.933-3.582 2.5-8 2.5s-8-.567-8-2.5S7.582 14 12 14s8 2.567 8 4.5Z"/></g><g><path fill="currentColor" opacity=".05" d="M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0"/><path stroke="currentColor" stroke-width="2" d="M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0Z"/></g></svg>`)
	case "profiles":
		return lucide.SquareUser()
	}
	return nil
}

var (
	settingsCategoryIcon_general      settingsIcon
	settingsCategoryIcon_shell        settingsIcon
	settingsCategoryIcon_remote       settingsIcon
	settingsCategoryIcon_controls     settingsIcon
	settingsCategoryIcon_editor       settingsIcon
	settingsCategoryIcon_integrations settingsIcon
	settingsCategoryIcon_view         settingsIcon
	settingsCategoryIcon_appearance   settingsIcon
	settingsCategoryIcon_recipes      settingsIcon
	settingsCategoryIcon_keybindings  settingsIcon
	settingsCategoryIcon_data         settingsIcon
	settingsCategoryIcon_advanced     settingsIcon
	settingsCategoryIcon_account      settingsIcon
)
