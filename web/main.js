function setupTable(rows, prefix) {
  const table = $("#example").DataTable({
    data: rows,
    ordering: false,
    order: [[1, "desc"]],
    columns: [
      {
        data: "title",
        render: function (data, type, row) {
          const link = row.is_paid
            ? prefix + row.post_id
            : "https://medium.com/articles/" + row.post_id;
          return `<div class="row">
          <a href=${link}>
            <h5>${row.title}</a> ${
            row.is_paid == 0
              ? ""
              : `<img src="paywall-svgrepo-com.svg" width="16" height="16"/>`
          }</h5>
          </div>
          <div class="row">
            <div class="col-auto">
              <img src="calendar-arrow-up-svgrepo-com.svg" width="16" height="16"/><small> ${formatDate(
                row.published_at
              )}</small>
            </div>
            <div class="col-auto">
              <small>${
                row.collection
                  ? `<img src="collection-svgrepo-com.svg" width="16" height="16"/> ${row.collection}`
                  : ""
              }</small>
            </div>
            <div class="col-auto">
            <small>${row.author ? `by ${row.author}` : ""}</small>
            </div>
          </div>
          <div class="row">
            <div class="col-auto">
              <img src="clap-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
                row.total_clap_count
              )}
            </div>
            <div class="col-auto">
              <img src="time-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
                row.reading_time
              )}
            </div>
            <div class="col-auto">
              <img src="share-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
                row.recommend_count
              )}
            </div>
          <div class="col-auto">
            <img src="comment-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
              row.response_count
            )}
          </div>
          <div class="row">
          <small>${row.tags ? row.tags.split(",").join(", ") : ""}</small>
          </div>
          <div class="row">
          <div class="col-auto">
          <button id="close" type="button" class="btn btn-link">
          <img src="close-svgrepo-com.svg" width="24" height="24"/>
          </button>
          </div>
          <div class="col-auto">
          <button id="share" type="button" class="btn btn-link">
          <img src="share-ios-export-svgrepo-com.svg" width="24" height="24"/>
          </button>
          </div>
          <div class="col-auto">
          <button id="check" type="button" class="btn btn-link">
          <img src="check-svgrepo-com.svg" width="24" height="24"/>
          </button>
          </div>
          </div>
        </div>`;
        },
      },
    ],
  });

  table.on("click", "button", function (e) {
    const data = table.row(e.target.closest("tr")).data();
    handleButton(e.currentTarget.id, data, link);
  });

  table.on("touchend", "button", function (e) {
    const data = table.row(e.target.closest("tr")).data();
    handleButton(e.currentTarget.id, data, link);
  });
  return table;
}
async function main() {
  const res = await fetch("https://medium-crawler.enzosv.workers.dev");
  const data = (await res.json()).results;
  const disliked = [];
  const liked = [];
  const def = [];
  const shared = [];
  for (const d of data) {
    const toggle = localStorage.getItem(d.post_id);
    if (toggle == 1) {
      liked.push(d);
    } else if (toggle == -1) {
      disliked.push(d);
    } else if (toggle == 0) {
      shared.push(d);
    } else {
      def.push(d);
    }
  }
  const freedium = window.location.href.includes("freedium");
  const prefix = freedium
    ? "https://freedium.cfd/"
    : "https://medium.com/articles/";
  const table = setupTable(def, prefix);

  $("#default").on("click", function () {
    table.clear();
    table.rows.add(def);
    table.draw();
  });
  $("#liked").on("click", function () {
    table.clear();
    table.rows.add(liked);
    table.draw();
  });
  $("#disliked").on("click", function () {
    table.clear();
    table.rows.add(disliked);
    table.draw();
  });
  $("#shared").on("click", function () {
    table.clear();
    table.rows.add(shared);
    table.draw();
  });
  $("#all").on("click", function () {
    table.clear();
    table.rows.add(data);
    table.draw();
  });
}

function handleButton(id, data, link) {
  switch (id) {
    case "close":
      localStorage.setItem(data.post_id, -1);
      return;
    case "share":
      share(data.title, link);
      localStorage.setItem(data.post_id, 0);
      return;
    case "check":
      localStorage.setItem(data.post_id, 1);
      return;
  }
}

function formatDate(date) {
  const month = [
    "Jan",
    "Feb",
    "Mar",
    "Apr",
    "May",
    "Jun",
    "Jul",
    "Aug",
    "Sep",
    "Oct",
    "Nov",
    "Dec",
  ];
  const d = new Date(date);
  return month[d.getUTCMonth()] + " " + d.getFullYear();
}

function share(title, link) {
  if (navigator.share) {
    navigator.share({
      title: title,
      url: link,
    });
    return;
  }
  if (navigator.clipboard) {
    navigator.clipboard.writeText(link);
    return;
  }
  console.log(navigator);
}

function cleanNumber(number) {
  if (number > 1000) {
    return Math.round(number / 1000) + "k";
  }
  return Math.round(number);
}

$("document").ready(function () {
  main();
});
