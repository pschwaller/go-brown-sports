package pkg

// AutomationMarker This is used to identify calendar events created by this service.
// It allows us to ignore other events that may be on the same calendar.
const AutomationMarker = "\n\nCreated by go-brown-sports automation.\n"

// SpreadsheetId The Google spreadsheet ID.  This is contained within the URL of the spreadsheet when viewing as a user.
// TODO This should be stored in a config file or as command line input.
const SpreadsheetId = "1j_0dCDYpTAgbgJzTfq_SrWXmVTkCMDRpKKWkgfYiPa8"

// CalendarId The calendar ID of the Google calendar we want to synchronize.  This can be found on the properties of the calendar.
// For the primary calendar of a user, specify "primary".
const CalendarId = "c_eb497c05a37742f18ad84c070a681e61f8bd70d6c0c49c8193d82f9b26106619@group.calendar.google.com"
