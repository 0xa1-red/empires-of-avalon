CREATE TABLE snapshots (
    kind varchar(255) not null,
    identity uuid not null,
    data bytea not null,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


/**
    Name:      House,
    BuildTime: "10s",
    Cost: []*ResourceCost{
        {Resource: Wood, Amount: 20, Permanent: true},
        {Resource: Population, Amount: 2, Permanent: false},
    },
    Generators: []blueprints.Generator{
        {
            Name:       string(Population),
            Amount:     1,
            TickLength: "2s",
        },
    },
    Limits: map[ResourceName]int{
        Population: 6,
    },
**/

CREATE TABLE blueprints (
    id uuid not null,
    kind varchar(255) not null,
    data json not null,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

