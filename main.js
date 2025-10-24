function diffDate(expDate) {
  var currentDate = new Date().valueOf();
  var expDateObj = new Date(expDate).valueOf();
  console.log(`Current unix date is ${currentDate}`);
  console.log(`Expiry unix date is ${expDateObj}`)
  var diff = expDateObj - currentDate;
  var absDiff = Math.abs(diff);
  if (absDiff != diff) {
    return "EXPIRED"
  }
  console.log(`Unix date diff is ${absDiff}`);
  var timeLeft = humanizeDuration(new Date(absDiff).valueOf(), {units: ["d", "h"], round: true});
  console.log(`Time left calc'd at ${timeLeft}`);
  return timeLeft;
};


// Render the table
function loadTable() {
  fetch("progs.json")
  .then((response) => response.json())
  .then((data) => {
    progs = data.progs;
    progs.sort((a, b) => {
      return new Date(a.expired_at) - new Date(b.expired_at)
    })
    const table = document.getElementById("main-table");
    progs.forEach((p) => {
      const listing = document.createElement("tr");
      listing.innerHTML = `
      <th scope="row">${diffDate(p.expired_at)}</th>
        <td><a href="${p.prog_url}" class="link-info">${p.prog_name}</a></td>
        <td><a href="${p.ep_url}" class="link-warning">${p.ep_name}</a></td>
      `;
      table.appendChild(listing);
    });
    const txtbox = document.getElementById("lu-textbox");
    txtbox.innerHTML = "Last updated: " + data.last_updated;
  })
  .catch((error) => console.error("Error: ", error));
}
