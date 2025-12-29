SELECT 
    a.id, a.title, a.slug, a.content, a.status, 
    a.author_id, a.category_id, a.published_at, a.created_at, a.updated_at,
    t2.id AS tag_id, t2.name AS tag_name, t2.slug AS tag_slug, t2.created_at AS tag_created_at
FROM articles a
INNER JOIN article_tags at_filter ON a.id = at_filter.article_id AND at_filter.tag_id = ?
LEFT JOIN article_tags at ON a.id = at.article_id
LEFT JOIN tags t2 ON at.tag_id = t2.id
WHERE a.status = 'published'
ORDER BY a.published_at DESC, t2.id ASC

