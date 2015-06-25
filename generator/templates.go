package generator

var indexTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
{{#.}}
  <sitemap>
    <loc>{{.}}</loc>
  </sitemap>
{{/.}}
</sitemapindex>
`

var mapTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
  xmlns:xhtml="http://www.w3.org/1999/xhtml">
  {{#.}}
  <url>
    <loc>{{Loc}}</loc>
    {{#Alternates}}
    <xhtml:link 
      rel="alternate"
      hreflang="{{HrefLang}}"
      href="{{Href}}"
    />
    {{/Alternates}}
  </url>
  {{/.}}
</urlset>
`
