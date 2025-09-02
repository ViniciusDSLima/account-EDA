package account

// Repository define a interface para operações de persistência com a entidade Account
type Repository interface {
	Save(account *Account) error
	FindByID(id string) (*Account, error)
	FindByEmail(email string) (*Account, error)
	FindAll() ([]*Account, error)
	Update(account *Account) error
	Delete(id string) error
}
