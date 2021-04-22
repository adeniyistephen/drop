package studio

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nextwavedevs/drop/business/auth"
	"github.com/nextwavedevs/drop/business/validate"
	"github.com/nextwavedevs/drop/foundation/database"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrNotFound is used when a specific Studio is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// Studio manages the set of API's for user access.
type Studio struct {
	log *log.Logger
	db  *mongo.Client
}

// New constructs a User for api access.
func New(log *log.Logger, db *mongo.Client) Studio {
	return Studio{
		log: log,
		db:  db,
	}
}

var studioCollection *mongo.Collection = database.OpenCollection(database.Client, "studio") // get collection "users" from db() which returns *mongo.Client

func (u Studio) Create(ctx context.Context, traceID string, ns NewStudio, now time.Time) (Info, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.studio.create")
	defer span.End()
	
	if err := validate.Check(ns); err != nil {
		return Info{}, errors.Wrap(err, "validating data")
	}

	std := Info{
		ID:           validate.GenerateID(),
		Name:         ns.Name,
		Email:        ns.Email,
		SocialHandle: ns.SocialHandle,
		City:         ns.City,
		Description:  ns.Description,
		State:        ns.State,
		Country:      ns.Country,
		Created_at:   now.UTC(),
	}

	insertResult, err := studioCollection.InsertOne(ctx, std)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult)
	return std, nil
}

// Update replaces a user document in the database.
func (u Studio) Update(ctx context.Context, traceID string, studioID string, us UpdateStudio, now time.Time) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.studio.update")
	defer span.End()

	if err := validate.CheckID(studioID); err != nil {
		return ErrInvalidID
	}
	if err := validate.Check(us); err != nil {
		return errors.Wrap(err, "validating data")
	}

	std, err := u.QueryByID(ctx, traceID, studioID)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	filter := bson.D{{Key: "name", Value: us.Name}} // converting value to BSON type
	after := options.After                // for returning updated document
	returnOpt := options.FindOneAndUpdateOptions{

		ReturnDocument: &after,
	}
	update := bson.M{
			"$set": bson.M{
			"name": us.Name,
			"email": us.Email,
			"socials": us.SocialHandle,
			"description": us.Description,
			"city": us.City,
			"state": us.State,
			"country": us.Country,
			},
		}
	updateResult := studioCollection.FindOneAndUpdate(ctx, filter, update, &returnOpt)

	var result Info
	_ = updateResult.Decode(&result)

	u.log.Printf("%s: %s", traceID, "studio.Update")
	fmt.Println("updated studio with id: ",std.ID)
	return nil
}

// Delete removes a user from the database.
func (u Studio) Delete(ctx context.Context, traceID string, claims auth.Claims, studioID string) error {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.studio.delete")
	defer span.End()

	if err := validate.CheckID(studioID); err != nil {
		return ErrInvalidID
	}

	// If you are not an admin and looking to delete someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) {
		return ErrForbidden
	}

	opts := options.Delete().SetCollation(&options.Collation{}) // to specify language-specific rules for string comparison, such as rules for lettercase
	res, err := studioCollection.DeleteOne(ctx, bson.D{{Key:"_id", Value: studioID}}, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted %v documents\n", res.DeletedCount)
	u.log.Printf("%s: %s", traceID, "studio.Delete")

	return nil
}

// Query retrieves a list of existing users from the database.
func (u Studio) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]*Info, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.studio.query")
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
	cur, err := studioCollection.Find(ctx, bson.D{{}},findOptionsOffset,findOptionPage) //returns a *mongo.Cursor
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
	u.log.Printf("%s: %s", traceID, "studio.Query")

	return results, nil
}

// QueryByID gets the specified user from the database.
func (u Studio) QueryByID(ctx context.Context, traceID string, studioID string) (Info, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.studio.querybyid")
	defer span.End()

	if err := validate.CheckID(studioID); err != nil {
		return Info{}, ErrInvalidID
	}

	var result Info //  an unordered representation of a BSON document which is a Map
	err := studioCollection.FindOne(ctx, bson.D{{Key:"_id",Value: studioID}}).Decode(&result)
	if err != nil {

		fmt.Println(err)

	}
	u.log.Printf("%s: %s", traceID, "studio.QueryByID")

	return result, nil
}

// Query retrieves a list of existing users from the database.
func (u Studio) QueryByLocation(ctx context.Context, traceID string, pageNumber int, rowsPerPage int, city string) ([]*Info, error) {

	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.studio.querybylocation")
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

	var results []*Info //slice for multiple documents
	cur, err := studioCollection.Find(ctx, bson.D{{Key:"city",Value: city}},findOptionsOffset,findOptionPage) //returns a *mongo.Cursor
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
	u.log.Printf("%s: %s", traceID, "studio.QueryByLocation")

	return results, nil
}