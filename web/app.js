async function main() {
  const sqlPromise = initSqlJs({
    locateFile: (file) => `https://sql.js.org/dist/${file}`,
  });

  const dataPromise = fetch("../medium.db").then((res) => res.arrayBuffer());
  const [SQL, buf] = await Promise.all([sqlPromise, dataPromise]);
  const db = new SQL.Database(new Uint8Array(buf));
  const stmt = db.prepare(
    `SELECT title, total_clap_count claps, 
    'https://medium.com/articles/' || post_id link, 
    date(published_at/1000, 'unixepoch') publish_date,
    c.name collection, 
    recommend_count, response_count, reading_time, tags
    FROM posts p
    LEFT OUTER JOIN collections c
        ON c.collection_id = p.collection
    ORDER BY total_clap_count DESC;`
  );
  const rows = [];
  while (stmt.step()) {
    const row = stmt.getAsObject();
    rows.push(row);
  }
  $("#example").DataTable({
    data: rows,
    order: [[1, "desc"]],
    columns: [
      {
        data: "title",
        render: function (data, type, row) {
          return `<div>
          <a href=${row.link}>${row.title}</a><br>
          ${row.collection ? `<subtitle>in ${row.collection}<br>` : ""}
          <img src="calendar-arrow-up-svgrepo-com.svg" width="16" height="16"/> ${
            row.publish_date
          }<br>
          <img src="clap-svgrepo-com.svg" width="16" height="16"/> ${row.claps}
          <img src="time-svgrepo-com.svg" width="16" height="16"/> ${Math.round(
            row.reading_time
          )}
          <img src="share-svgrepo-com.svg" width="16" height="16"/> ${
            row.recommend_count
          }
          <img src="comment-svgrepo-com.svg" width="16" height="16"/> ${
            row.response_count
          }
          </subtitle>
          </div>`;
        },
      },
    ],
  });
}

main();
