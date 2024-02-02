go-brown-sports
===============

A program written in Go to synchronize the Brown Game Day Worker schedule 
with a Google calendar.

# Overview
The Brown Athletics communications staff coordinates a set of workers
for a variety of roles at Brown sporting events.  These roles include 
public address announcer, statisticians, photographers, and 
videographers.

For people who are frequently on the schedule, referring to the 
spreadsheet for their assignments is practical and has low overhead. 
For people who are scheduled less than once a week, it makes more sense
to manually create entries on their own personal calendars.  Depending upon
the number of events, this can be a tedious and error-prone process.

This program seeks to automate the maintenance of a Google calendar
that is synchronized with the content of the main scheduling spreadsheet.

# Design Details

## Spreadsheet Structure

This program makes a few assumptions based on the current structure of
the worker schedule spreadsheet:
* There are separate tabs for each month, named with the full month name (e.g., January instead of Jan)
* The first three columns of each month tab are "Date", "Time", and "Sport"
* The Date column is formatted as "Monday, January 2, 2006". If this format is changed, this program will need to change accordingly.
* The Time column is in AM/PM times. This program parses most variants that have been observed thus far.
* Columns D and above are named roles. Role names can be changed and roles can be added without making any changes to this program.
* An "x" or a blank in a role cell is ignored.

Additionally, there is a Worker Contact Info tab with the name and contact 
information for all workers.

## Synchronization

When the program runs, it generates a list of all future events from the 
spreadsheet, and compared it to the future events on the calendar. 

The spreadsheet is considered the reliable source-of-truth.
If an event appears on the spreadsheet but not the calendar, 
it is added. If an event
appears on the calendar but not the spreadsheet it is deleted. 
If changes to roles or workers are made, those changes are updated in the existing
calendar entry.

Past events 
are ignored and not changed.


## Accessing the Calendar

The public web browser view of the calendar is [HERE](https://calendar.google.com/calendar/embed?src=c_eb497c05a37742f18ad84c070a681e61f8bd70d6c0c49c8193d82f9b26106619%40group.calendar.google.com&ctz=America%2FNew_York).

# Future Updates and Enhancements

## Update Schedule

* Need to determine how frequently the program should synchronize the calendar
from the spreadsheet.
* Need to determine where the program should be hosted.  Default is on a Linux 
host in the author's basement.

## Invites to Worker's Calendars

With a few hours of coding and testing, the program can be enhanced to 
determine the email address of each worker, and send them a calendar invite.
This invite would be updated if changes are made, either to the start time or 
to the other workers at that event.

Options for choosing who gets an invite:
* All workers get an invite
* An additional column is added to the Workers Contact Info tab. This would need
to be updated by someone with write-access.
* Create a Google Form that allows workers to set their own opt-in status. Using the "collect email"
feature of Google Forms ensures that only the account owner can change the opt-in status. 

## Coordination with Athletics Communications

* The program has a mapping of "Sport" to a Location string that is used
in the calendar invite. This mapping is currently incomplete.  Estimated to 
need about 10 minutes of someone's time.
* Once the "Invites to Worker's Calendars" feature is implemented, there are a 
number of people scheduled for Roles who do not show up in the Workers Contact Info tab.
This would need to be updated.  Estimated to need about 30 minutes of someone's time.

# Running the Code
## Credentials
In order to run the code you must first get a credentials.json file in the current directory.
Follow the steps at the [Quickstart](https://developers.google.com/sheets/api/quickstart/go)
to create the file.

After OAuth2 flows, the program will create a token.json file in the current 
directory.