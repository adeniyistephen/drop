package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/nextwavedevs/drop/business/auth"
	"github.com/nextwavedevs/drop/business/validate"
	"github.com/nextwavedevs/drop/foundation/database"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// User manages the set of API's for user access.
type User struct {
	log *log.Logger
	db  *mongo.Client
}

// New constructs a User for api access.
func New(log *log.Logger, db *mongo.Client) User {
	return User{
		log: log,
		db:  db,
	}
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user") // get collection "users" from db() which returns *mongo.Client

// Create inserts a new user into the database.
func (u User) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (Info, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.create")
	defer span.End()

	if err := validate.Check(nu); err != nil {
		return Info{}, errors.Wrap(err, "validating data")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return Info{}, errors.Wrap(err, "generating password hash")
	}

	usr := Info{
		ID:           validate.GenerateID(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Password:     nu.Password,
		Roles:        nu.Roles,
		Created_at:   now.UTC(),
		Updated_at:   now.UTC(),
	}

	insertResult, err := userCollection.InsertOne(ctx, usr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult)
	u.log.Printf("%s: %s", traceID, "user.Create")
	return usr, nil
}

// Update replaces a user document in the database.
func (u User) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.update")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return ErrInvalidID
	}
	if err := validate.Check(uu); err != nil {
		return errors.Wrap(err, "validating data")
	}

	usr, err := u.QueryByID(ctx, traceID, claims, userID)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		usr.PasswordHash = pw
	}
	usr.Updated_at = now

	filter := bson.D{{Key: "name", Value: uu.Name}} // converting value to BSON type
	after := options.After                // for returning updated document
	returnOpt := options.FindOneAndUpdateOptions{

		ReturnDocument: &after,
	}
	update := bson.M{
			"$set": bson.M{
			"name": *uu.Name,
			"email": *uu.Email,
			"roles": uu.Roles,
			"password": *uu.Password,
			"password_hash": usr.PasswordHash,
			},
		}
	updateResult := userCollection.FindOneAndUpdate(ctx, filter, update, &returnOpt)

	usr.Updated_at = now

	var result Info
	_ = updateResult.Decode(&result)
	u.log.Printf("%s: %s", traceID, "user.Update")

	return nil
}

// Delete removes a user from the database.
func (u User) Delete(ctx context.Context, traceID string, claims auth.Claims, userID string) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.delete")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return ErrInvalidID
	}

	// If you are not an admin and looking to delete someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return ErrForbidden
	}

	opts := options.Delete().SetCollation(&options.Collation{}) // to specify language-specific rules for string comparison, such as rules for lettercase
	res, err := userCollection.DeleteOne(ctx, bson.D{{Key:"_id", Value: userID}}, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted %v documents\n", res.DeletedCount)
	u.log.Printf("%s: %s", traceID, "user.Delete")

	return nil
}

// Query retrieves a list of existing users from the database.
func (u User) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]*Info, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.query")
	defer span.End()

	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	// Pass these options to the Find method
	findOptionsOffset := options.Find()
	findOptionPage := options.Find()
	findOptionsOffset.SetLimit(int64(data.Offset))
	findOptionPage.SetLimit(int64(data.RowsPerPage))

	var results []*Info                                   //slice for multiple documents
	cur, err := userCollection.Find(ctx, bson.D{{}},findOptionsOffset,findOptionPage) //returns a *mongo.Cursor
	if err != nil {
		fmt.Println(err)
	}
	for cur.Next(ctx) { //Next() gets the next document for corresponding cursor

		var elem Info
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, &elem) // appending document pointed by Next()
	}
	cur.Close(ctx) // close the cursor once stream of documents has exhausted
	fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
	u.log.Printf("%s: %s", traceID, "user.Query")

	return results, nil
}

// QueryByID gets the specified user from the database.
func (u User) QueryByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (Info, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyid")
	defer span.End()

	if err := validate.CheckID(userID); err != nil {
		return Info{}, ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return Info{}, ErrForbidden
	}

	var result Info //  an unordered representation of a BSON document which is a Map
	err := userCollection.FindOne(ctx, bson.D{{Key:"_id",Value: userID}}).Decode(&result)
	if err != nil {

		fmt.Println(err)

	}
	u.log.Printf("%s: %s", traceID, "user.QueryByID")

	return result, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (u User) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.authenticate")
	defer span.End()

	data := struct {
		Email string `db:"email"`
	}{
		Email: email,
	}

	var usr Info
	err := userCollection.FindOne(ctx, bson.D{{Key:"email",Value: data.Email}}).Decode(&usr)
	if err != nil {
		fmt.Println(err)
	}
	u.log.Printf("%s: %s", traceID, "user.Authenticate")
	
	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return auth.Claims{}, database.ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "drop project",
			Subject:   usr.ID,
			ExpiresAt: jwt.At(now.Add(time.Hour)),
			IssuedAt:  jwt.At(now),
		},
		Roles: usr.Roles,
	}

	return claims, nil
}