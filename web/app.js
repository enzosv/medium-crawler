async function main() {
  let data = await fetch("./medium.csv").then((response) => response.text());
  data = data.split("\n").map((v) => v.split(","));
  $("#example").DataTable({
    data: data,
    order: [[1, "desc"]],
    columns: [
      {
        data: "title",
        render: function (data, type, row) {
          return `<div>
          <a href=https://medium.com/articles/${row[2]}>${row[0].replaceAll(
            "|",
            ","
          )}</a><br>
          ${row[4] ? `<subtitle>in ${row[4]}` : ""}
          <img src="calendar-arrow-up-svgrepo-com.svg" width="16" height="16"/> ${
            row[3]
          }<br>
          <img src="clap-svgrepo-com.svg" width="16" height="16"/> ${row[1]}
          <img src="time-svgrepo-com.svg" width="16" height="16"/> ${row[7]}
          <img src="share-svgrepo-com.svg" width="16" height="16"/> ${row[5]}
          <img src="comment-svgrepo-com.svg" width="16" height="16"/> ${
            row[6]
          }<br>
          </subtitle>
          <a style="text-decoration: none; display: flex; align-items: center;" tabindex="-1" aria-label="Omnivore logo" href="https://omnivore.app/api/save?url=https://freedium.cfd/${
            row[2]
          }">
          </div>`;
        },
      },
    ],
  });
}

main();
