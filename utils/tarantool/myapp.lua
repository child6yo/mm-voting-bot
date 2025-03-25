box.schema.space.create('votings', {if_not_exists = true})

box.space.votings:format({
    { name = 'id', type = 'number' },
    { name = 'user_id', type = 'number' },
})

box.space.votings:create_index('primary', { parts = { 'id' } })
box.space.votings:create_index('user_id', { parts = { 'user_id' } })