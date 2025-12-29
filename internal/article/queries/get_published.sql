SELECT 
    a.id, a.title, a.slug, a.content, a.status, 
    a.author_id, a.category_id, a.published_at, a.created_at, a.updated_at,
    t.id AS tag_id, t.name AS tag_name, t.slug AS tag_slug, t.created_at AS tag_created_at
FROM articles a
LEFT JOIN article_tags at ON a.id = at.article_id
LEFT JOIN tags t ON at.tag_id = t.id
WHERE a.status = 'published'
ORDER BY a.published_at DESC, t.id ASC

