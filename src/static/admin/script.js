domain = 'https://crowd.fastforward.team/'
function parseJwt (token) {
    var base64Url = token.split('.')[1];
    var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    var jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(jsonPayload);
};

function regenTokens() {
    if (localStorage.getItem('reftoken') === null) {
        window.location.replace("login");
    }
    let r = localStorage.getItem('reftoken');
    fetch(domain+'admin/api/newacctoken', {
    method: "post",
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      reftoken: r
       })
    })
      .then(res => res.json())
      .then(data => {
        sessionStorage.setItem('acctoken', data.acctoken);
        sessionStorage.setItem('username', parseJwt(data.acctoken).aud);
        $('#username-span').text(sessionStorage.getItem('username'));
      })
      .catch(error => {
        Swal.fire({
          toast: true,
          position: 'top',
          showConfirmButton: false,
          title: 'Failed to verify login details',
          icon: 'error',
          timer: 2000,
          timerProgressBar: true,
          didDestroy: () => {
            window.location.replace("login");
        }
        })
      })
}
function getReported() {
    let bearer = 'Bearer ' + sessionStorage.getItem('acctoken')
    let page='1'
    fetch(domain + 'admin/api/getreported?page=' + page, {
        method: 'post',
        headers: {
            'Authorization': bearer,
            'Content-Type': 'application/x-www-form-url-encoded' 
        }
    })
    .then (res => {
        console.log(res.text)
        if (res.headers.get('content-type') == 'text/plain; charset=utf-8') {
            $('#table').text('No reported links so far')
        }
    })
    .then(res => res.json())
    .then(data => {
        data.forEach(function(obj) { 
            obj.link = obj.domain + "/" + obj.path
            delete obj.domain
            delete obj.path
        makeTable(data)
     })
    })
    .catch(error => ({
        message: 'Something bad happened ' + error
    }))
}
regenTokens()
setInterval(function(){
    regenTokens()
    }, 870000);
getReported()
var linkTable
function makeTable(data) {
    linkTable = new gridjs.Grid({
        columns: [{
            id: 'link',
            name: 'Link',
            formatter: (_, row) => gridjs.html(`<a href='https://${row.cells[0].data}'>${row.cells[0].data}</a>`)
        }, {
            id: 'destination',
            name: gridjs.html('Submitted<br>Target'),

        }, {
            id: 'times_reported',
            name: gridjs.html(' Times<br> Reported'),
        }, {
            id: 'voted_by',
            name: gridjs.html('Voted for <br>deletion by'),
            width: '30%'
        }, { 
            name: gridjs.html('Vote Target <br> for deletion'),
            formatter: (_, row) => {
              return gridjs.h('button', {
                onClick: () => voteDelete(row.cells[0].data)
              }, 'Vote');
            },
         }
        ],
        data: data,
        autoWidth: true
      }).render(document.getElementById("table"));
}

function refreshTable() {
    let bearer = 'Bearer ' + sessionStorage.getItem('acctoken')
    let page='1'
    fetch(domain + 'admin/api/getreported?page=' + page, {
        method: 'post',
        headers: {
            'Authorization': bearer,
            'Content-Type': 'application/x-www-form-url-encoded' 
        }
    })
    .then(res => res.json())
    .then(data => {
        data.forEach(function(obj) { 
            obj.link = obj.domain + "/" + obj.path
            delete obj.domain
            delete obj.path
        renderTable(data)
     })
    })
    .catch(error => ({
        message: 'Something bad happened ' + error
    }))
}

function renderTable(data) {
    linkTable.updateConfig({
        data: data
    }).forceRender();
}

function voteDelete(link) {
domain=link.substr(0,link.indexOf('/')); 
path=link.substr(link.indexOf('/')+1);
let bearer = 'Bearer ' + sessionStorage.getItem('acctoken')
fetch(domain + 'admin/api/votedelete', {
method: "post",
headers: {
    'Authorization': bearer,
    'Accept': 'application/json',
    'Content-Type': 'application/json'
},
body: JSON.stringify({
  domain: domain,
  path: path
  })
})
  .then(res => {
    if (res.status == 200) {
      Swal.fire({
        toast: true,
        position: 'bottom-end',
        showConfirmButton: false,
        title: 'Successfully voted to delete',
        icon: 'success',
        timer: 2000,
      })
      refreshTable()
    } else if (res.status == 202) {
        Swal.fire({
            toast: true,
            position: 'bottom-end',
            showConfirmButton: false,
            title: 'Successfully deleted submisson. [Reason: 2 votes]',
            icon: 'success',
            timer: 2000,
          })
        refreshTable()
    } else {
      Swal.fire({
        toast: true,
        position: 'top',
        showConfirmButton: false,
        title: 'Submisson was already deleted',
        icon: 'info',
        timer: 2000,
        timerProgressBar: true,
      })
    }
    })
}

