package deltas

import (
	"fmt"
	"testing"

	"github.com/beaconsoftwarellc/gadget/v2/database"
	dberrors "github.com/beaconsoftwarellc/gadget/v2/database/errors"
	"github.com/beaconsoftwarellc/gadget/v2/database/lock"
	"github.com/beaconsoftwarellc/gadget/v2/database/transaction"
	"github.com/beaconsoftwarellc/gadget/v2/errors"
	"github.com/beaconsoftwarellc/gadget/v2/generator"
	assert1 "github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func GetMocks(t *testing.T) (*database.MockConnection, *database.MockClient, *database.MockAPI,
	*transaction.MockTransaction, *transaction.MockImplementation) {
	ctrl := gomock.NewController(t)
	connection := database.NewMockConnection(ctrl)
	client := database.NewMockClient(ctrl)
	connection.EXPECT().Client().Return(client).AnyTimes()
	api := database.NewMockAPI(ctrl)
	tx := transaction.NewMockTransaction(ctrl)
	api.EXPECT().GetTransaction().Return(tx).AnyTimes()
	txImp := transaction.NewMockImplementation(ctrl)
	tx.EXPECT().Implementation().Return(txImp).AnyTimes()
	return connection, client, api, tx, txImp
}

func TestExecute_LockError(t *testing.T) {
	assert := assert1.New(t)
	connection, client, _, _, _ := GetMocks(t)
	deltas := []*Delta{
		{
			ID:     generator.Int(),
			Name:   generator.String(10),
			Script: generator.String(128),
		},
	}
	// lock acquisition
	expected := generator.String(32)
	client.EXPECT().Select(gomock.Any(),
		"SELECT GET_LOCK('delta_exec', 0) AS STATUS").Return(
		errors.New(expected))
	connection.EXPECT().Close().Return(nil)
	actual := execute(database.InstanceConfig{
		DeltaLockMaxTries:     1,
		DeltaLockMinimumCycle: 1,
		DeltaLockMaxCycle:     2,
	}, connection, "", deltas)
	assert.EqualError(actual, expected)
}

type lockMatcher struct {
}

func (lockMatcher) Matches(x interface{}) bool {
	rows, ok := x.(*[]*lock.StatusResult)
	if !ok {
		return false
	}
	*rows = []*lock.StatusResult{{Status: 1}}
	return true
}

func (lockMatcher) String() string {
	return "lock matcher"
}

func TestExecute_BeginError(t *testing.T) {
	assert := assert1.New(t)
	connection, client, api, _, _ := GetMocks(t)
	deltas := []*Delta{
		{
			ID:     generator.Int(),
			Name:   generator.String(10),
			Script: generator.String(128),
		},
	}
	// lock acquisition
	client.EXPECT().Select(&lockMatcher{}, "SELECT GET_LOCK('delta_exec', 0) AS STATUS").Return(nil)
	// lock release
	client.EXPECT().Select(gomock.Any(), "SELECT RELEASE_LOCK('delta_exec') AS STATUS").Return(nil)
	expected := generator.String(32)
	connection.EXPECT().Database().Return(api)
	api.EXPECT().Begin().Return(errors.New(expected))
	connection.EXPECT().Close().Return(nil)
	actual := execute(database.InstanceConfig{
		DeltaLockMaxTries:     1,
		DeltaLockMinimumCycle: 1,
		DeltaLockMaxCycle:     2,
	}, connection, "", deltas)
	assert.EqualError(actual, expected)
}

type tableExistsMatcher struct {
	tableExists bool
}

func (matcher *tableExistsMatcher) Matches(x interface{}) bool {
	rows, ok := x.(*[]*database.TableNameResult)
	if !ok {
		return false
	}
	if matcher.tableExists {
		*rows = []*database.TableNameResult{{}}
	}
	return true
}

func (matcher *tableExistsMatcher) String() string {
	return fmt.Sprintf("tableExistsMatcher(%v)", matcher.tableExists)
}

func TestExecute_TableExists_Error(t *testing.T) {
	assert := assert1.New(t)
	connection, client, api, _, _ := GetMocks(t)
	deltas := []*Delta{
		{
			ID:     generator.Int(),
			Name:   generator.String(10),
			Script: generator.String(128),
		},
	}
	schema := generator.String(32)
	// lock acquisition
	client.EXPECT().Select(&lockMatcher{}, "SELECT GET_LOCK('delta_exec', 0) AS STATUS").Return(nil)
	// lock release
	client.EXPECT().Select(gomock.Any(), "SELECT RELEASE_LOCK('delta_exec') AS STATUS").Return(nil)

	connection.EXPECT().Database().Return(api)
	api.EXPECT().Begin().Return(nil)
	tableExistsQuery := fmt.Sprintf(database.TableExistenceQueryFormat, schema, DeltaTableName)

	expected := generator.String(32)
	client.EXPECT().Select(&tableExistsMatcher{true}, tableExistsQuery).
		Return(errors.New(expected))
	connection.EXPECT().Close().Return(nil)
	actual := execute(database.InstanceConfig{
		DeltaLockMaxTries:     1,
		DeltaLockMinimumCycle: 1,
		DeltaLockMaxCycle:     2,
	}, connection, schema, deltas)
	assert.EqualError(actual, expected)
}

func TestExecute_TableDoesNotExist(t *testing.T) {
	assert := assert1.New(t)
	connection, client, api, _, txImp := GetMocks(t)
	deltas := []*Delta{
		{
			ID:     generator.Int(),
			Name:   generator.String(10),
			Script: generator.String(128),
		},
	}
	schema := generator.String(32)
	// lock acquisition
	client.EXPECT().Select(&lockMatcher{}, "SELECT GET_LOCK('delta_exec', 0) AS STATUS").Return(nil)
	// lock release
	client.EXPECT().Select(gomock.Any(), "SELECT RELEASE_LOCK('delta_exec') AS STATUS").Return(nil)

	connection.EXPECT().Database().Return(api)
	api.EXPECT().Begin().Return(nil)
	tableExistsQuery := fmt.Sprintf(database.TableExistenceQueryFormat, schema, DeltaTableName)

	client.EXPECT().Select(&tableExistsMatcher{false}, tableExistsQuery).Return(nil)

	txImp.EXPECT().Exec(CreateDeltaTableSQL).Return(nil, nil)

	// ExecuteDelta calls
	api.EXPECT().ReadOneWhere(gomock.Any(), DeltaMeta.ID.Equal(deltas[0].ID)).
		Return(dberrors.NewNotFoundError())
	txImp.EXPECT().Exec(deltas[0].Script).Return(nil, nil)
	api.EXPECT().Create(&DeltaRecord{ID: deltas[0].ID, Name: deltas[0].Name}).Return(nil)
	// / ExecuteDelta calls

	api.EXPECT().Commit().Return(nil)
	connection.EXPECT().Close().Return(nil)

	actual := execute(database.InstanceConfig{
		DeltaLockMaxTries:     1,
		DeltaLockMinimumCycle: 1,
		DeltaLockMaxCycle:     2,
	}, connection, schema, deltas)
	assert.NoError(actual)
}

func TestExecute(t *testing.T) {
	assert := assert1.New(t)
	connection, client, api, _, txImp := GetMocks(t)
	deltas := []*Delta{
		{
			ID:     generator.Int(),
			Name:   generator.String(10),
			Script: generator.String(128),
		},
	}
	schema := generator.String(32)
	// lock acquisition
	client.EXPECT().Select(&lockMatcher{}, "SELECT GET_LOCK('delta_exec', 0) AS STATUS").Return(nil)
	// lock release
	client.EXPECT().Select(gomock.Any(), "SELECT RELEASE_LOCK('delta_exec') AS STATUS").Return(nil)

	connection.EXPECT().Database().Return(api)
	api.EXPECT().Begin().Return(nil)
	tableExistsQuery := fmt.Sprintf(database.TableExistenceQueryFormat, schema, DeltaTableName)

	client.EXPECT().Select(&tableExistsMatcher{true}, tableExistsQuery).Return(nil)

	// ExecuteDelta calls
	api.EXPECT().ReadOneWhere(gomock.Any(), DeltaMeta.ID.Equal(deltas[0].ID)).
		Return(dberrors.NewNotFoundError())
	txImp.EXPECT().Exec(deltas[0].Script).Return(nil, nil)
	api.EXPECT().Create(&DeltaRecord{ID: deltas[0].ID, Name: deltas[0].Name}).Return(nil)
	// / ExecuteDelta calls

	api.EXPECT().Commit().Return(nil)
	connection.EXPECT().Close().Return(nil)

	actual := execute(database.InstanceConfig{
		DeltaLockMaxTries:     1,
		DeltaLockMinimumCycle: 1,
		DeltaLockMaxCycle:     2,
	}, connection, schema, deltas)
	assert.NoError(actual)
}

func TestExecute_Rollback(t *testing.T) {
	assert := assert1.New(t)
	connection, client, api, _, txImp := GetMocks(t)
	deltas := []*Delta{
		{
			ID:     generator.Int(),
			Name:   generator.String(10),
			Script: generator.String(128),
		},
	}
	schema := generator.String(32)
	// lock acquisition
	client.EXPECT().Select(&lockMatcher{}, "SELECT GET_LOCK('delta_exec', 0) AS STATUS").Return(nil)
	// lock release
	client.EXPECT().Select(gomock.Any(), "SELECT RELEASE_LOCK('delta_exec') AS STATUS").Return(nil)

	connection.EXPECT().Database().Return(api)
	api.EXPECT().Begin().Return(nil)
	tableExistsQuery := fmt.Sprintf(database.TableExistenceQueryFormat, schema, DeltaTableName)

	client.EXPECT().Select(&tableExistsMatcher{true}, tableExistsQuery).Return(nil)

	// ExecuteDelta calls
	api.EXPECT().ReadOneWhere(gomock.Any(), DeltaMeta.ID.Equal(deltas[0].ID)).
		Return(dberrors.NewNotFoundError())
	txImp.EXPECT().Exec(deltas[0].Script).Return(nil, nil)
	expected := generator.String(32)
	api.EXPECT().Create(&DeltaRecord{ID: deltas[0].ID, Name: deltas[0].Name}).
		Return(errors.New(expected))
	// / ExecuteDelta calls

	api.EXPECT().Rollback().Return(nil)
	connection.EXPECT().Close().Return(nil)

	actual := execute(database.InstanceConfig{
		DeltaLockMaxTries:     1,
		DeltaLockMinimumCycle: 1,
		DeltaLockMaxCycle:     2,
	}, connection, schema, deltas)
	assert.EqualError(actual, expected)
}

func TestExecuteDelta_AlreadyExecuted(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	api := database.NewMockAPI(ctrl)
	delta := &Delta{
		ID:   generator.Int(),
		Name: generator.String(32),
	}
	api.EXPECT().ReadOneWhere(gomock.Any(), DeltaMeta.ID.Equal(delta.ID)).Return(nil)
	err := ExecuteDelta(api, delta)
	assert.NoError(err)
}

func TestExecuteDelta_UnexpectedError(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	api := database.NewMockAPI(ctrl)
	delta := &Delta{
		ID:   generator.Int(),
		Name: generator.String(32),
	}
	expected := generator.String(32)
	api.EXPECT().ReadOneWhere(gomock.Any(), DeltaMeta.ID.Equal(delta.ID)).
		Return(errors.New(expected))
	actual := ExecuteDelta(api, delta)
	assert.EqualError(actual, expected)
}

func TestExecuteDelta_ExecError(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	api := database.NewMockAPI(ctrl)
	tx := transaction.NewMockTransaction(ctrl)
	txImp := transaction.NewMockImplementation(ctrl)
	delta := &Delta{
		ID:     generator.Int(),
		Name:   generator.String(32),
		Script: generator.String(128),
	}
	expected := generator.String(32)
	api.EXPECT().ReadOneWhere(gomock.Any(), DeltaMeta.ID.Equal(delta.ID)).
		Return(dberrors.NewNotFoundError())
	api.EXPECT().GetTransaction().Return(tx)
	tx.EXPECT().Implementation().Return(txImp)
	txImp.EXPECT().Exec(delta.Script).Return(nil, errors.New(expected))
	actual := ExecuteDelta(api, delta)
	assert.EqualError(actual, expected)
}

func TestExecuteDelta(t *testing.T) {
	assert := assert1.New(t)
	ctrl := gomock.NewController(t)
	api := database.NewMockAPI(ctrl)
	tx := transaction.NewMockTransaction(ctrl)
	txImp := transaction.NewMockImplementation(ctrl)
	delta := &Delta{
		ID:     generator.Int(),
		Name:   generator.String(32),
		Script: generator.String(128),
	}
	expected := generator.String(32)
	api.EXPECT().ReadOneWhere(gomock.Any(), DeltaMeta.ID.Equal(delta.ID)).
		Return(dberrors.NewNotFoundError())
	api.EXPECT().GetTransaction().Return(tx)
	tx.EXPECT().Implementation().Return(txImp)
	txImp.EXPECT().Exec(delta.Script).Return(nil, nil)
	api.EXPECT().Create(&DeltaRecord{ID: delta.ID, Name: delta.Name}).
		Return(errors.New(expected))
	actual := ExecuteDelta(api, delta)
	assert.EqualError(actual, expected)
}
