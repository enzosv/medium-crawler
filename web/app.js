async function main() {
  let data = await fetch("./medium.csv").then((response) => response.text());
  data = data.split("\n").map((v) => v.split(","));
  data.pop(); // remove newline at end
  $("#example").DataTable({
    data: data,
    ordering: false,
    order: [[1, "desc"]],
    columns: [
      {
        data: "title",
        render: function (data, type, row) {
          return `<div class="row">
          <a href=https://medium.com/articles/${row[2]}>
          <h6>${row[0].replaceAll("|", ",")}</h6></a>
          <div>
          ${row[4] ? `in ${row[4]}` : ""}
          <img src="calendar-arrow-up-svgrepo-com.svg" width="16" height="16"/> ${
            row[3]
          }<br>
          <img src="clap-svgrepo-com.svg" width="16" height="16"/> ${row[1]}
          <img src="time-svgrepo-com.svg" width="16" height="16"/> ${row[7]}
          <img src="share-svgrepo-com.svg" width="16" height="16"/> ${row[5]}
          <img src="comment-svgrepo-com.svg" width="16" height="16"/> ${
            row[6]
          }<br>
          <a href="https://omnivore.app/api/save?url=https://freedium.cfd/https://medium.com/articles/${
            row[2]
          }"></a>
          ${row[8] ? row[8].replaceAll("|", ", ") : ""}
          </div>
          </div>`;
        },
      },
    ],
  });
}

main();
