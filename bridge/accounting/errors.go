package accounting

import "errors"

var (
	ErrIDNotExist                = errors.New("this id doesn't exist")
	ErrIDAlreadyExists           = errors.New("this id already exsits")
	ErrWorkersAreNotEnough       = errors.New("There are not enough workers")
	ErrDataPoolHolderNotExist    = errors.New("this user's datapool doesn't exist")
	ErrDataPoolAlreadyExists     = errors.New("this user's datapool already exists")
	ErrCreateDataPoolNotComplete = errors.New("failed to complete to create datepool")
	ErrArgumentIsInvalid         = errors.New("an argument is invalid")
	ErrFailedToComplete          = errors.New("Failed to complete work")
)
