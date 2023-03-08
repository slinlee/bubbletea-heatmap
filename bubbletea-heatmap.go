package bubbleteaheatmap

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	selectedX int
	selectedY int
	calData   []CalDataPoint
	viewData  [52][7]viewDataPoint // Hardcoded to one year for now
	// focus     bool // TODO
}

var scaleColors = []string{
	// Light Theme
	// #ebedf0 - Less
	// #9be9a8
	// #40c463
	// #30a14e
	// #216e39 - More

	// Dark Theme
	"#161b22", // Less
	"#0e4429",
	"#006d32",
	"#26a641",
	"#39d353", // - More

}

type CalDataPoint struct {
	Date  time.Time
	Value float64
}

func (m Model) addCalData(date time.Time, val float64) {
	// Create new cal data point and add to cal data
	newPoint := CalDataPoint{date, val}
	m.calData = append(m.calData, newPoint)
}

func getIndexDate(x int, y int) time.Time {
	// compare the x,y to today and subtract
	today := time.Now()
	todayX, todayY := getDateIndex(today)

	diffX := todayX - x
	diffY := todayY - y

	diffDays := diffX*7 + diffY

	targetDate := today.AddDate(0, 0, -diffDays)
	return targetDate
}

func weeksAgo(date time.Time) int {
	today := truncateToDate(time.Now())
	thisWeek := today.AddDate(0, 0, -int(today.Weekday())) // Most recent Sunday

	compareDate := truncateToDate(date)
	compareWeek := compareDate.AddDate(0, 0, -int(compareDate.Weekday()))

	result := thisWeek.Sub(compareWeek).Hours() / 24 / 7
	return int(result)
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Local().Year(), t.Local().Month(), t.Local().Day(), 0, 0, 0, 0, t.Local().Location())
}

func getDateIndex(date time.Time) (int, int) {
	// Max index - number of weeks ago
	x := 51 - weeksAgo(date)

	y := int(date.Local().Weekday())

	return x, y
}

func parseCalToView(calData []CalDataPoint) [52][7]viewDataPoint {
	var viewData [52][7]viewDataPoint

	for _, v := range calData {
		x, y := getDateIndex(v.Date)
		// Check if in range
		// TODO: un-hardcode the X limit
		if x > -1 && y > -1 &&
			x < 52 && y < 7 {
			viewData[x][y].actual += v.Value
		}
	}
	viewData = normalizeViewData(viewData)
	return viewData
}

func normalizeViewData(data [52][7]viewDataPoint) [52][7]viewDataPoint {
	var min float64
	var max float64

	// Find min/max
	min = data[0][0].actual
	max = data[0][0].actual

	for _, row := range data {
		for _, val := range row {

			if val.actual < min {
				min = val.actual
			}
			if val.actual > max {
				max = val.actual
			}
		}
	}

	// Normalize the data
	for i, row := range data {
		for j, val := range row {
			data[i][j].normalized = (val.actual - min) / (max - min)
		}
	}
	return data
}

type viewDataPoint struct {
	actual     float64
	normalized float64
}

func getScaleColor(value float64) string {
	const numColors = 5
	// Assume it's normalized between 0.0-1.0
	const max = 1.0
	const min = 0.0

	return scaleColors[int((value/max)*(numColors-1))]
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Create a new model with default settings.
func New(data []CalDataPoint) Model {
	todayX, todayY := getDateIndex(time.Now())

	parsedData := parseCalToView(data)
	return Model{
		selectedX: todayX,
		selectedY: todayY,
		calData:   data,
		viewData:  parsedData,
		// focus:     false, // TODO
	}
}

// func (m *Model) Focus() tea.Cmd { // TODO
// m.focus = true
// return m.Cursor.Focus()
// }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// TODO: ignore if not focused
	// if !m.focus { return m, nil }

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedY > 0 {
				m.selectedY--
			} else if m.selectedY == 0 && m.selectedX > 0 {
				// Scroll to the end of the previous week if on Sunday
				// and not at the beginning of the calendar.
				m.selectedY = 6
				m.selectedX--
			}

		case "down", "j":
			// Don't allow user to scroll beyond today
			if m.selectedY < 6 &&
				(m.selectedX != 51 ||
					m.selectedY < int(time.Now().Weekday())) {
				m.selectedY++
			} else if m.selectedY == 6 && m.selectedX != 51 {
				// Scroll to the beginning of next week if on Saturday
				// and not at the end of the calendar.
				m.selectedY = 0
				m.selectedX++
			}
		case "right", "l":
			// Don't allow users to scroll beyond today from the previous column
			if m.selectedX < 50 ||
				(m.selectedX == 50 &&
					m.selectedY <= int(time.Now().Weekday())) {
				m.selectedX++
			}
		case "left", "h":
			if m.selectedX > 0 {
				m.selectedX--
			}
		case "enter", " ":
			// Hard coded to add a new entry with value `1.0`
			m.addCalData(
				getIndexDate(m.selectedX, m.selectedY),

				1.0)
			m.viewData = parseCalToView(m.calData)

		}
	}
	return m, nil
}

func (m Model) View() string {
	// The header

	theTime := getIndexDate(m.selectedX, m.selectedY) // time.Now()

	title, _ := glamour.Render(theTime.Format("# Monday, January 02, 2006"), "dark")
	s := title

	selectedDetail := "    Value: " +
		fmt.Sprint(m.viewData[m.selectedX][m.selectedY].actual) +
		" normalized: " +
		fmt.Sprint(m.viewData[m.selectedX][m.selectedY].normalized) +
		"\n\n"

	s += selectedDetail

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	boxStyle := lipgloss.NewStyle().
		PaddingRight(1).
		Foreground(lipgloss.Color(scaleColors[2]))

	boxSelectedStyle := boxStyle.Copy().
		Background(lipgloss.Color("#9999ff")).
		Foreground(lipgloss.Color(scaleColors[0]))

	// Month Labels
	var currMonth time.Month
	s += "  "
	for j := 0; j < 52; j++ {
		// Check the last day of the week for that column
		jMonth := getIndexDate(j, 6).Month()

		if currMonth != jMonth {
			currMonth = jMonth
			s += labelStyle.Render(getIndexDate(j, 6).Format("Jan") + " ")

			// Skip the length of the label we just added
			j += 1
		} else {
			s += "  "
		}
	}
	s += "\n"

	for j := 0; j < 7; j++ {
		// Add day of week labels
		switch j {
		case 0:
			s += labelStyle.Render("S ")
		case 1:
			s += labelStyle.Render("M ")
		case 2:
			s += labelStyle.Render("T ")
		case 3:
			s += labelStyle.Render("W ")
		case 4:
			s += labelStyle.Render("T ")
		case 5:
			s += labelStyle.Render("F ")
		case 6:
			s += labelStyle.Render("S ")
		}

		// Add calendar days
		for i := 0; i < 52; i++ {
			// Selected Item
			if m.selectedX == i && m.selectedY == j {
				s += boxSelectedStyle.Copy().Foreground(
					lipgloss.Color(
						getScaleColor(
							m.viewData[i][j].normalized))).
					Render("■")
			} else if i == 51 &&
				j > int(time.Now().Weekday()) {

				// In the future
				s += boxStyle.Render(" ")
			} else {

				// Not Selected Item and not in the future
				s += boxStyle.Copy().
					Foreground(
						lipgloss.Color(
							getScaleColor(
								m.viewData[i][j].normalized))).
					Render("■")
			}
		}
		s += "\n"
	}

	return s
}
