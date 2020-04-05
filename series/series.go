package series

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// seriesStartDate is our default start date
var seriesStartDate = time.Date(2020, 1, 22, 0, 0, 0, 0, time.UTC)

// NewData returns a new Data series based on the row values
// We expect the cols
func NewData(row []string) (*Data, error) {

	country := row[0]
	province := row[1]
	areaID, err := strconv.Atoi(row[2])
	if err != nil {
		return nil, fmt.Errorf("areas: invalid population at row:%s", row)
	}

	latitude, err := strconv.ParseFloat(row[3], 64)
	if err != nil {
		return nil, fmt.Errorf("areas: invalid latitude at row:%s", row)
	}
	longitude, err := strconv.ParseFloat(row[4], 64)
	if err != nil {
		return nil, fmt.Errorf("areas: invalid longitude at row:%s", row)
	}
	population, err := strconv.Atoi(row[5])
	if err != nil {
		return nil, fmt.Errorf("areas: invalid population at row:%s", row)
	}

	var lockdown time.Time
	if row[6] != "" {
		lockdown, err = time.Parse("2006-01-02", row[6])
		if err != nil {
			return nil, fmt.Errorf("areas: invalid lockdown at row:%s", row)
		}
	}

	color := row[7]

	// NB updated at left at zero time
	s := &Data{
		ID:         areaID,
		Country:    country,
		Province:   province,
		Latitude:   latitude,
		Longitude:  longitude,
		Population: population,
		Color:      color,
		LockdownAt: lockdown,
		Days:       make([]*Day, 0),
	}

	return s, nil
}

// Data stores data for one country or province within a country
type Data struct {

	// The arbitrary unique numeric id of this area - used as foreign key in series tables
	ID int

	// The Country or Region
	Country string

	// The Province or State - blank for countries
	Province string

	// The population of the area (if known)
	Population int

	// Coordinates for this area
	Latitude, Longitude float64

	// An rgb color/colour for plotting charts
	Color string

	// UTC Date data last updated
	UpdatedAt time.Time

	// UTC Date full area lockdown started
	LockdownAt time.Time

	// Days containing all our data - each day holds cumulative totals
	Days []*Day
}

// Format formats a given number for display and returns a string
func (d *Data) Format(i int) string {
	if i < 10000 {
		return fmt.Sprintf("%d", i)
	} else if i < 1000000 {
		return fmt.Sprintf("%.1fk", float64(i)/1000)
	}

	return fmt.Sprintf("%.3gm", float64(i)/1000000)
}

// Global returns true if this is the global series
func (d *Data) String() string {
	if d.IsGlobal() {
		return fmt.Sprintf("%s (%d)", "Global", len(d.Days))
	} else if d.Province == "" {
		return fmt.Sprintf("%s (%d)", d.Country, len(d.Days))
	}
	return fmt.Sprintf("%s, %s (%d)", d.Province, d.Country, len(d.Days))
}

// Title returns a display title for this series
func (d *Data) Title() string {
	if d.IsGlobal() {
		return "Global"
	} else if d.IsCountry() {
		return d.Country
	}

	return fmt.Sprintf("%s (%s)", d.Province, d.Country)
}

// SetUpdated updates UpdatedAt if it is before this new time
func (d *Data) SetUpdated(updated time.Time) {
	if d.UpdatedAt.Before(updated) {
		d.UpdatedAt = updated
	}
}

// IsGlobal returns true if this is the global series
func (d *Data) IsGlobal() bool {
	return d.Country == "" && d.Province == ""
}

// IsEuropean returns true if this is a European country
func (d *Data) IsEuropean() bool {
	if d.Province != "" {
		return false
	}
	return d.Country == "United Kingdom" || d.Country == "France" || d.Country == "Italy" || d.Country == "Belgium" || d.Country == "Spain" || d.Country == "Germany" || d.Country == "Netherlands" || d.Country == "Switzerland" || d.Country == "Sweden" || d.Country == "Portugal"
}

// IsCountry returns true if this is the global series
func (d *Data) IsCountry() bool {
	return !d.IsGlobal() && !d.IsProvince()
}

// IsProvince returns true if this is a province under a country
func (d *Data) IsProvince() bool {
	return d.Country != "" && d.Province != ""
}

// Valid returns true if this series is valid
// a series without days is considered invalid
func (d *Data) Valid() bool {
	return len(d.Days) == 0
}

// Key converts a value into one suitable for use in urls
func (d *Data) Key(v string) string {
	return strings.Replace(strings.ToLower(v), " ", "-", -1)
}

// Match returns true if this series matches country and province
// performs a case insensitive match
func (d *Data) Match(country string, province string) bool {
	return d.MatchCountry(country) && d.MatchProvince(province)
}

// MatchCountry return true if this series matches country
// performs a case insensitive match
func (d *Data) MatchCountry(country string) bool {
	return d.Key(d.Country) == d.Key(country)
}

// MatchProvince return true if this series matches province
// performs a case insensitive match
func (d *Data) MatchProvince(province string) bool {
	return d.Key(d.Province) == d.Key(province)
}

// FetchDate returns the datapoint for a given date and dataKind
func (d *Data) FetchDate(date time.Time, dataKind int) int {

	for _, d := range d.Days {
		if d.Date.Equal(date) {
			switch dataKind {
			case DataDeaths:
				return d.Deaths
			case DataConfirmed:
				return d.Confirmed
			case DataRecovered:
				return d.Recovered
			case DataTested:
				return d.Tested
			}
		}
	}

	return 0
}

// Period returns a subset of the series data just for the no of days specified
func (d *Data) Period(days int) *Data {
	// for now just return all days
	return d
}

// LastDay returns the last day in the series
// a blank day is returned if no days
func (d *Data) LastDay() *Day {
	if len(d.Days) == 0 {
		return &Day{}
	}
	return d.Days[len(d.Days)-1]
}

// PenultimateDay returns the second last day in the series
// a blank day is returned if no days
func (d *Data) PenultimateDay() *Day {
	if len(d.Days) < 2 {
		return &Day{}
	}
	return d.Days[len(d.Days)-2]
}

// TotalDeaths returns the cumulative death due to COVID-19 for this series
func (d *Data) TotalDeaths() int {
	return d.LastDay().Deaths
}

// TotalConfirmed returns the cumulative confirmed cases of COVID-19 for this series
func (d *Data) TotalConfirmed() int {
	return d.LastDay().Confirmed
}

// TotalRecovered returns the cumulative recovered cases of COVID-19 for this series
func (d *Data) TotalRecovered() int {
	return d.LastDay().Recovered
}

// TotalTested returns the cumulative tested cases of COVID-19 for this series
func (d *Data) TotalTested() int {
	return d.LastDay().Tested
}

// DeathsToday returns deaths for last day in series - day before
func (d *Data) DeathsToday() int {
	return d.LastDay().Deaths - d.PenultimateDay().Deaths
}

// ConfirmedToday returns confirmed for last day in series - day before
func (d *Data) ConfirmedToday() int {
	return d.LastDay().Confirmed - d.PenultimateDay().Confirmed
}

// Deaths returns cumulative totals of deaths as integer values
func (d *Data) Deaths() (values []int) {
	for _, day := range d.Days {
		values = append(values, day.Deaths)
	}
	return values
}

// Confirmed returns cumulative totals of confirmed as integer values
func (d *Data) Confirmed() (values []int) {
	for _, day := range d.Days {
		values = append(values, day.Confirmed)
	}
	return values
}

// DeathsDaily returns an array of int values for deaths per day
func (d *Data) DeathsDaily() (values []int) {
	var previous int
	for _, day := range d.Days {
		values = append(values, day.Deaths-previous)
		previous = day.Deaths
	}
	return values
}

// ConfirmedDaily returns an array of int values for confirmed per day
func (d *Data) ConfirmedDaily() (values []int) {
	var previous int
	for _, day := range d.Days {
		values = append(values, day.Confirmed-previous)
		previous = day.Confirmed
	}
	return values
}

// DaysFrom returns day counts from a series of numbers
func (d *Data) DaysFrom(values []int) []string {

	// Build a set of labels for these values counting from day 1
	var days []string
	for i := range values {
		days = append(days, fmt.Sprintf("Day %d", i+1))
	}
	return days
}

// DeathsFrom returns series after death number n
func (d *Data) DeathsFrom(n int) []int {
	// Walk through deaths looking for death n, then return series from that day
	for i, day := range d.Days {
		if day.Deaths >= n {
			return d.Deaths()[i : len(d.Days)-1]
		}
	}
	return nil
}

// AverageDeaths returns the average deaths per day over the last 3 days
func (d *Data) AverageDeaths() int {
	// If not enough days, return 0
	if len(d.Days) < 3 {
		return 0
	}

	// Get deaths over last 3 days
	sum := d.Days[len(d.Days)-1].Deaths - d.Days[len(d.Days)-3].Deaths

	// return simple average
	return sum / 3
}

// AverageConfirmed returns the average confirmed per day over the last 3 days
func (d *Data) AverageConfirmed() int {
	// If not enough days, return 0
	if len(d.Days) < 3 {
		return 0
	}

	// Get deaths over last 3 days
	sum := d.Days[len(d.Days)-1].Confirmed - d.Days[len(d.Days)-3].Confirmed

	// return simple average
	return sum / 3
}

// DoubleDeathDays returns the number of days it took to more than double deaths
// this ignores today's incomplete data
func (d *Data) DoubleDeathDays() (days int) {
	i := d.Count() - 1
	half := d.Days[i].Deaths / 2
	for i--; i >= 0; i-- {
		if d.Days[i].Deaths < half {
			break
		}
		days++
	}
	// Return the number of days required to halve count
	return days
}

// DoubleConfirmedDays returns the number of days it took to more than double confirmed
// this ignores today's incomplete data
func (d *Data) DoubleConfirmedDays() (days int) {
	i := d.Count() - 1
	half := d.Days[i].Confirmed / 2
	for i--; i >= 0; i-- {
		if d.Days[i].Confirmed < half {
			break
		}
		days++
	}
	// Return the number of days required to halve count
	return days
}

// LastHours returns the number of hours that have passed since 0 UTC
func (d *Data) LastHours() int {
	return time.Now().UTC().Hour()
}

// Dates returns a set of date labels as an array of strings
// for every datapoint in this series
func (d *Data) Dates() (dates []string) {
	for _, day := range d.Days {
		dates = append(dates, day.Date.Format("Jan 2"))
	}
	return dates
}

// Count returns the count of days in this series
func (d *Data) Count() int {
	return len(d.Days)
}

// SetDayData sets the data for a given day,
// the day should be added first with AddDays if required
func (d *Data) SetDayData(dayNo, deaths, confirmed, recovered, tested int) error {
	index := dayNo - 1
	if index > len(d.Days)-1 {
		return fmt.Errorf("series: index out of range for set day:%d len:%d", index, len(d.Days))
	}

	day := d.Days[index]
	day.SetAllData(deaths, confirmed, recovered, tested)
	return nil
}

// SetData adds the given series of data to this series
// existing data for that dataKind will be replaced
func (d *Data) SetData(startDate time.Time, dataKind int, values []int) error {

	//log.Printf("data: set data of kind:%d data:%v", dataKind, values)

	// If we don't have enough days, add some
	if len(d.Days) < len(values) {
		//log.Printf("addDays:%d %d", len(d.Days), len(values))
		d.AddDays(len(values) - len(d.Days))
	}

	// Now set the values for this datakind on each day we have
	for i, day := range d.Days {

		// Check date on first day matches
		if i == 0 && !day.Date.Equal(startDate) {
			return fmt.Errorf("series: mismatch on start date for data:%v %v", startDate, day.Date)
		}
		//log.Printf("day:%d", values[i])
		// Fill in the value on each day from values
		err := day.SetData(dataKind, values[i])
		if err != nil {
			return fmt.Errorf("series: failed to add day:%v error:%s", day, err)
		}
	}

	return nil
}

// MergeData adds the given series of data to this series
// existing data for that dataKind will have these values added
func (d *Data) MergeData(startDate time.Time, dataKind int, values []int) error {

	if false {
		log.Printf("data: merge data of kind:%d data:%v", dataKind, values)
	}

	// If we don't have enough days, add some
	if len(d.Days) < len(values) {
		//log.Printf("addDays:%d %d", len(d.Days), len(values))
		d.AddDays(len(values) - len(d.Days))
	}

	// Now set the values for this datakind on each day we have
	for i, day := range d.Days {

		// Check date on first day matches
		if i == 0 && !day.Date.Equal(startDate) {
			return fmt.Errorf("series: mismatch on start date for data:%v %v", startDate, day.Date)
		}
		//log.Printf("day:%d", values[i])
		// Fill in the value on each day from values
		err := day.MergeData(dataKind, values[i])
		if err != nil {
			return fmt.Errorf("series: failed to add day:%v error:%s", day, err)
		}
	}

	return nil
}

// MergeSeries will merge the data from the incoming series with this one
// start dates are assumed to be the same
func (d *Data) MergeSeries(series *Data) error {

	// Add days if required
	if len(d.Days) < len(series.Days) {
		//log.Printf("addDays:%d", len(series.Days)-len(d.Days))
		d.AddDays(len(series.Days) - len(d.Days))
	}

	//log.Printf("days:%d sdays:%d", len(d.Days), len(series.Days))

	// Now add this dataset on top of ours using MergeDay
	for i, day := range d.Days {

		// Fill in the value on each day from values
		// if dates don't match an error will be returned
		// check we have this day first in series - we silenty ignore too many days
		if i < len(series.Days) {
			err := day.MergeDay(series.Days[i])
			if err != nil {
				return fmt.Errorf("series: failed to add day:%v error:%s", day, err)
			}
		} else {
			//	log.Printf("days overflow on series:%s", series)
		}

	}

	return nil
}

// AddDays adds the given number of days to the end of our series
func (d *Data) AddDays(count int) {
	// Get the last day (if any), and start a day after, otherwise start afresh
	date := seriesStartDate
	if len(d.Days) > 0 {
		date = d.LastDay().Date.AddDate(0, 0, 1)
	}

	for i := 0; i < count; i++ {
		day := &Day{
			Date: date,
		}
		d.Days = append(d.Days, day)
		date = date.AddDate(0, 0, 1)
	}
}

// FIXME - I think this won't be required

// AddDay adds a day to this series
// an error is returned if the date is not at the end of the series
func (d *Data) AddDay(date time.Time, deaths, confirmed, recovered, tested int) error {
	// Check data is valid
	if date.IsZero() {
		return fmt.Errorf("series: invalid zero date in AddDay")
	}

	// Check date is more than the last date in series
	if len(d.Days) > 0 {
		if !d.LastDay().Date.Before(date) {
			return fmt.Errorf("series: invalid date added")
		}
	}

	// What about updating an existing day, do we ever do that?
	// Different function for that.

	day := &Day{
		Date:      date,
		Deaths:    deaths,
		Confirmed: confirmed,
		Recovered: recovered,
		Tested:    tested,
	}

	d.Days = append(d.Days, day)
	return nil
}