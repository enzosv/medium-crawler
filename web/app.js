const is_omnivore = window.location.href.includes("omnivore");

function buildLink(url) {
  if (is_omnivore) {
    return `<a href="javascript:void(0);" onclick="omnivore('${url}')">`;
    // return `<a href=https://omnivore.app/api/save?url=${url}`;
  }
  return `<a href=${url}>`;
}

async function main() {
  const res = await fetch("./medium.json");
  const data = await res.json();
  const freedium = window.location.href.includes("freedium");
  const prefix =
    freedium || is_omnivore
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
            ${buildLink(prefix + row[2])}
            <h5>${row[0].replaceAll("|", ",")}</a> ${
            row[9] == 0
              ? ""
              : `<img src="paywall-svgrepo-com.svg" width="16" height="16"/>`
          }</h5>
          </div>
          <div class="row">
            <div class="col">
              <img src="calendar-arrow-up-svgrepo-com.svg" width="16" height="16"/><small>${
                row[3]
              }</small>
            </div>
            <div class="col">
              <small>${row[4] ? `in ${row[4]}` : ""}</small>
            </div>
            <div class="col">
              <small>${row[10] ? `by ${row[10]}` : ""}</small>
            </div>
          </div>
          <div class="row">
            <div class="col">
              <img src="clap-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
                row[1]
              )}
            </div>
            <div class="col">
              <img src="time-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
                row[7]
              )}
            </div>
            <div class="col">
              <img src="share-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
                row[5]
              )}
            </div>
          <div class="col">
            <img src="comment-svgrepo-com.svg" width="16" height="16"/> ${cleanNumber(
              row[6]
            )}
          </div>
          <div class="row">
          <small>${row[8] ? row[8] : ""}</small>
          </div>
         
          ${
            freedium
              ? `
              <div class="row">
                <a title="Save to Omnivore" onclick="omnivore('${prefix}${row[2]}')" href="javascript:void(0);">
                <svg width="26" height="26" fill="none">
                <path d="M8.42285 17.9061V10.5447C8.42285 9.91527 9.16173 9.55951 9.65432 9.99737L11.9257 13.3087C12.3909 13.6918 13.0477 13.6918 13.5129 13.3087L15.7296 10.0247C16.2222 9.61424 16.961 9.94263 16.961 10.5721V14.458C16.961 16.3463 18.2199 17.8788 20.1081 17.8788H20.1629C21.7775 17.8788 23.1731 16.7841 23.5563 15.2243C23.7478 14.4033 23.912 13.5549 23.912 12.8982C23.8847 6.46715 18.4388 1.596 11.9257 2.03385C6.39776 2.41698 1.9371 6.87764 1.55397 12.4056C1.11612 18.9187 6.26093 24.3645 12.7193 24.3645" stroke="white" stroke-width="2.18182" stroke-miterlimit="10"></path>
                </svg></a>
              </div>`
              : ""
          }
        </div>`;
        },
      },
    ],
  });
}

function cleanNumber(number) {
  if (number > 1000) {
    return Math.round(number / 1000) + "k";
  }
  return Math.round(number);
}

async function omnivore(link) {
  key = localStorage.getItem("omnivore");
  if (!key) {
    key = prompt("omnivore api key");
    localStorage.setItem("omnivore", key);
  }
  const body = {
    query:
      "mutation SaveUrl($input: SaveUrlInput!) { saveUrl(input: $input) { ... on SaveSuccess { url clientRequestId } ... on SaveError { errorCodes message } } }",
    variables: {
      input: {
        source: "api",
        url: link,
        clientRequestId: crypto.randomUUID(),
        labels: [{ name: "medium-crawler" }],
      },
    },
  };

  const response = await fetch("https://api-prod.omnivore.app/api/graphql", {
    body: JSON.stringify(body),
    headers: {
      "Content-Type": "application/json",
      Authorization: localStorage.getItem("omnivore"),
    },
    method: "POST",
  });
  const data = await response.json();
  const url = data?.data?.saveUrl?.url;

  if (omnivore && url) {
    window.open(url);
    return;
  }
  const toastLiveExample = document.getElementById("liveToast");
  const toast = new bootstrap.Toast(toastLiveExample);
  const toastMessage = document.getElementById("toast");
  if (url) {
    toastMessage.innerHTML = `Saved! <a href=${url}>${url}</a>`;
  } else {
    console.log(response);
    toastMessage.innerHTML = JSON.stringify(data);
  }

  toast.show();
}
main();
