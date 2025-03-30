function filterTable() {
    const searchBox = document.getElementById('search');
    const table = document.getElementById("myTable");
    const trs = table.tBodies[0].getElementsByTagName("tr");
    const filter = searchBox.value.toUpperCase();

    if(filter !== " "){
        for (let rowI = 0; rowI < trs.length; rowI++) {
            // define the row's cells
            const tds = trs[rowI].getElementsByTagName("td");
            // hide the row
            trs[rowI].style.display = "none";
            // loop through row cells
            for (let cellI = 1; cellI < tds.length-1; cellI++) {
                // if there's a match
                if (tds[cellI].innerHTML.toUpperCase().indexOf(filter) > -1) {
                    // show the row
                    trs[rowI].style.display = "";
                    // skip to the next row
                    continue;
                }
            }
        }
    }
}

function showLoading(isShow) {
    if (isShow) {
        document.getElementById('loadingOverlay').style.display = 'flex';
    } else {
        document.getElementById('loadingOverlay').style.display = 'none';
    }
}