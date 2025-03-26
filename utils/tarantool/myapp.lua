box.space.answers:drop({if_exists = true})
box.space.votings:drop({if_exists = true})
box.sequence.votings_id:drop({if_exists = true})  
box.sequence.answers_id:drop({if_exists = true})

box.schema.sequence.create('votings_id', {if_not_exists = true})
box.schema.sequence.create('answers_id', {if_not_exists = true})

box.schema.space.create('votings', {
    format = {
        {name = 'id', type = 'unsigned'},
        {name = 'user_id', type = 'string'},
    }
})

box.space.votings:create_index('primary', {
    parts = {'id'},
    if_not_exists = true
})

box.schema.space.create('answers', {
    format = {
        {name = 'id', type = 'unsigned'},
        {name = 'voting_id', type = 'unsigned', foreign_key = {space = 'votings', field = 'id'}},
        {name = 'description', type = 'string'},
        {name = 'votes', type = 'number', default = 0}
    }
})

box.space.answers:create_index('primary', {
    parts = {{field = 'id', type = 'unsigned'}},
    if_not_exists = true
})

box.space.answers:create_index('voting_idx', {
    parts = {{field = 'voting_id', type = 'unsigned'}}, 
    if_not_exists = true
})

function create_voting_with_answers(user_id, answers)
    local voting_id = box.sequence.votings_id:next()
    
    box.space.votings:insert({
        voting_id,
        user_id,
    })
    
    for _, desc in ipairs(answers) do
        box.space.answers:insert({
            box.sequence.answers_id:next(),
            voting_id,
            desc,
            0
        })
    end
    
    return voting_id
end