async function main() {
  let data = await fetch("./medium.csv").then((response) => response.text());
  data = data.split("\n").map((v) => v.split(","));
  data.pop(); // remove newline at end
  const prefix = window.location.href.includes("freedium")
    ? "https://freedium.cfd/https://medium.com/articles/"
    : "https://medium.com/articles/";
  $("#example").DataTable({
    data: data,
    ordering: false,
    order: [[1, "desc"]],
    columns: [
      {
        data: "title",
        render: function (data, type, row) {
          return `<div class="row">
          <a href=${prefix}${row[2]}>
          <h6>${row[0].replaceAll("|", ",")}</a> ${
            row[9] == 0
              ? ""
              : `<img src="paywall-svgrepo-com.svg" width="16" height="16"/>`
          }</h6>
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
          <a href="https://omnivore.app/api/save?url=${prefix}${row[2]}"></a>
          ${row[8] ? row[8].replaceAll("|", ", ") : ""}
          </div>
          </div>`;
        },
      },
    ],
  });
}

// omnivore("https://freedium.cfd/https://medium.com/articles/ac8e5bc9b455");

// async function omnivore(link) {
//   const body = {
//     query:
//       "mutation SaveUrl($input: SaveUrlInput!) { saveUrl(input: $input) { ... on SaveSuccess { url clientRequestId } ... on SaveError { errorCodes message } } }",
//     variables: {
//       input: {
//         source: "api",
//         url: link,
//         clientRequestId: crypto.randomUUID(),
//       },
//     },
//   };

//   await fetch("https://api-prod.omnivore.app/api/graphql", {
//     body: JSON.stringify(body),
//     headers: {
//       "Content-Type": "application/json",
//       Authorization: "bf328018-34fb-40ae-9c53-233bd3412ac5",
//     },
//     method: "POST",
//   });
// }
main();
