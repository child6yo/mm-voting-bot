local datetime = require('datetime')

box.space.answers:drop({if_exists = true})
box.space.votings:drop({if_exists = true})
box.sequence.votings_id:drop({if_exists = true})  
box.sequence.answers_id:drop({if_exists = true})

box.schema.sequence.create('votings_id', {if_not_exists = true})
box.schema.sequence.create('answers_id', {if_not_exists = true})

box.schema.space.create('votings', {
    format = {
        {name = 'id', type = 'unsigned'},
        {name = 'user_id', type = 'string', unique = false},
        {name = 'expires_at', type = 'datetime'}
    }
})

box.space.votings:create_index('primary', {
    parts = {'id'},
    if_not_exists = true
})

box.schema.space.create('answers', {
    format = {
        {name = 'id', type = 'unsigned'},
        {name = 'voting_id', type = 'unsigned', foreign_key = {space = 'votings', field = 'id'}, unique = false},
        {name = 'local_id', type = 'unsigned', unique = false},
        {name = 'description', type = 'string', unique = false},
        {name = 'votes', type = 'number', default = 0, unique = false}
    }
})

box.space.answers:create_index('primary', {
    parts = {'id'},
    if_not_exists = true
})

box.space.answers:create_index('voting_idx', {
    parts = {{field = 'voting_id', type = 'unsigned'}},
    unique = false, 
    if_not_exists = true
})

function create_voting_with_answers(user_id, answers, duration_minutes)
    local now = datetime.now()
    duration = datetime.interval.new{min = duration_minutes}
    local expires_at = now:add(duration)

    local voting_id = box.sequence.votings_id:next()
    
    box.space.votings:insert({
        voting_id,
        user_id,
        expires_at
    })
    
    for i, desc in ipairs(answers) do
        box.space.answers:insert({
            box.sequence.answers_id:next(),
            voting_id,
            i,
            desc,
            0
        })
    end
    
    return voting_id
end

function delete_voting_with_answers(voting_id)
    box.space.answers.index.voting_idx:pairs{voting_id}:each(function(answer)
        box.space.answers:delete{answer.id}
    end)
    
    box.space.votings:delete{voting_id}
end