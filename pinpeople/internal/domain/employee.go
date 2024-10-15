package domain

import "time"

type Employee struct {
	ID                           int
	Name                         string
	Email                        string
	CorporateEmail               string
	DepartmentID                 *int
	Position                     *string
	Role                         *string
	LocationID                   *int
	TimeAtCompany                *string
	Gender                       *string
	Generation                   *string
	ResponseDate                 *time.Time
	PositionInterest             *int
	PositionInterestComments     *string
	Contribution                 *int
	ContributionComments         *string
	LearningDevelopment          *int
	LearningDevelopmentComments  *string
	Feedback                     *int
	FeedbackComments             *string
	ManagerInteraction           *int
	ManagerInteractionComments   *string
	CareerClarity                *int
	CareerClarityComments        *string
	RetentionExpectation         *int
	RetentionExpectationComments *string
	ENPS                         *int
	ENPSComments                 *string
	OpenENPS                     *string
}
