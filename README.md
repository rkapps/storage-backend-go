# storage-backend-go

A lightweight MongoDB wrapper for Go that simplifies common CRUD operations, enables Atlas Search out-of-the-box, and manages versioned schema migrations.

## 🚀 Key Features

1. Simplified CRUD: High-level wrappers for FindByID, Find, InsertOne, and InsertMany.
2. Atlas Search: Native support for full-text search aggregation stages.
3. Managed Migrations: Programmatic, version-controlled schema changes with "Up" and "Down" logic.
4. Type-Safe: Designed to work with native Go structs and BSON tags.

## Usage

### Initializing

    mongoConnStr := os.Getenv("MONGO_ATLAS_CONN_STR")
    dbName := os.Getenv("DB_NAME")

    client, err := mongodb.NewMongoClient(mongoConnStr, dbName)
    if err != nil {
        log.Fatalf("Error connecting to Mongo DB: %v", err)
    }

    err = migrations.RunMigrations(client)
    if err != nil {
        log.Fatal(err)
    }

### Defining the model

Create the Model and implement the interfaces Id(), SetId() and CollectionName()
    type UserModel struct {
        ID     string `bson:"id"`
        UserId string `bson:"user_id"`
        Name   string
        City   string
        Country string
    }

    func (u *UserModel) Id() string {
        return u.ID
    }

    func (u *UserModel) SetId() {
        u.ID = u.UserId
    }

    func (u *UserModel) CollectionName() string {
        return "user"
    }

### Migrations

In your migrations folder, create files with incremental version names and add the code in the init(). To use migrations you have to add the migrations schema first.

    001_add_migrations.go

    func init() {

        migrations.Register(1, "Migrations schema", func(client *mongodb.MongoClient) error {

            migrationColl := mongodb.NewMongoRepository[*migrations.Migration](*client)
            err := migrationColl.CreateIndexes(context.Background(), []mongo.IndexModel{
                {
                    Keys:    bson.D{{Key: "id", Value: 1}},
                    Options: options.Index().SetName("idx_id").SetUnique(true),
                },
                {
                    Keys:    bson.D{{Key: "version", Value: 1}},
                    Options: options.Index().SetName("idx_version").SetUnique(false),
                },
            })
            return err

        })
    }


    002_add_users_control.go

    func init() {
        migrations.Register(2, "User Control schema",
            func(client *mongodb.MongoClient) error {

                userColl := mongodb.NewMongoRepository[*UserControl](*client)
                err := userColl.CreateIndexes(context.Background(), []mongo.IndexModel{
                    {
                        Keys:    bson.D{{Key: "id", Value: 1}},
                        Options: options.Index().SetName("idx_id").SetUnique(true),
                    },
                })
                return err
            },
            func(client *mongodb.MongoClient) error {
                return nil
            },
        )
    }

### CRUD operations

1. FindById - Find a model by its unique id.

    ctx := context.Background()
    userRepo := NewMongoRepository[*UserModel](*client)
    user, err := userRepo.FindByID(ctx, "3")
    if err != nil {
        t.Errorf("Error finding user: %v :%v", "3", err)
    }

2. Find - Find used on a filter

    ctx := context.Background()
    userRepo := NewMongoRepository[*UserModel](*client)
    filter := bson.M{"city": bson.M{"$in": []string{"San Fransisco, Los Angeles, Sacrament}}}
    sort := bson.D{{Key: "name", Value: -1}}
    users, _ := userRepo.Find(ctx, filter, sort, 0, 0)

3. Atlas Search - Search uses Atlas search to find data. This requires a Atlas search index defined.

    criteria := mongodb.SearchCriteria{}
    criteria.IndexName = "idx_search"

    criteria.Query = "san"
    criteria.AutoCompleteFields = []string{"city", "country"}

    sortField := mongodb.CreateSortField("city", -1)
    criteria.SortFields = append(criteria.SortFields, sortField)

    data, err := tr.Search(ctx, criteria)

## 🤝 Contributing

Create a feature branch from main.
Ensure go test ./... passes.
Open a Pull Request with a description of your changes.
