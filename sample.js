redix.CreateIndex("user_interests", function (doc) {
    return CompoundIndex([doc.user_id, doc.category_id])
})

redix.CreateIndex("userPostsByUserID", function (doc) {
    return CompoundIndex([doc.user_id, doc.date])
})

redix.CreateIndex("userPostsByTags", function (doc) {
    return MultiIndex(_(doc.tags).map(function (tag) {
        return Compound([tag, doc.date])
    }))
})

redix.CreateAggregator("userCategoryInterests", function (doc) {
    return CounterAggregator()
})

/**

    mkidx compound user_interesets 'type == events' user_id
    mkdix multi user_posts 'type == posts'

*/