/* eslint-disable no-console */
/* eslint-env browser */
/* global Swal gridjs $ */

// Collection of functions to handle all the network requests, all funcs return a fetch promise
const netRequest = {
  accessToken: (refToken) => {
    const options = {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        reftoken: refToken,
      }),
    };
    return fetch('/admin/api/newacctoken', options);
  },

  reportedLinks(acctoken) {
    const bearer = `Bearer ${acctoken}`;
    const page = '1';
    const options = {
      method: 'POST',
      headers: {
        Authorization: bearer,
        'Content-Type': 'application/x-www-form-url-encoded',
      },
    };
    return fetch(`/admin/api/getreported?page=${page}`, options);
  },

  voteDelete(link, acctoken) {
    const bearer = `Bearer ${acctoken}`;
    const linkdomain = link.substr(0, link.indexOf('/'));
    const linkpath = link.substr(link.indexOf('/') + 1);
    const options = {
      method: 'POST',
      headers: {
        Authorization: bearer,
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        domain: linkdomain,
        path: linkpath,
      }),
    };
    return fetch('/admin/api/votedelete', options);
  },
};

// Clean netRequest wrapper, all funcs return a promise
const request = {

  // promise of returning an access token
  newAcessTkn: (reftoken) => new Promise((resolve, reject) => {
    netRequest.accessToken(reftoken)
      .then((response) => {
        if (response.status === 200) {
          response.json().then((data) => resolve(data));
        } else if (response.status === 401) {
          reject(new Error('unauthorized'));
        } else {
          reject(new Error(`something went wrong [${response}]`));
        }
      })
      .catch((error) => reject(error));
  }),

  // promise of returning a list of reported links
  getReported: (acctoken) => new Promise((resolve, reject) => {
    netRequest.reportedLinks(acctoken)
      .then((response) => {
        if (response.status === 204) {
          resolve([]);
        } else if (response.status === 200) {
          response.json().then((data) => {
            const dataMod = [];
            data.forEach((obj) => {
              const objMod = obj;
              objMod.link = `${obj.domain}/${obj.path}`;
              delete objMod.domain;
              delete objMod.path;
              objMod.delete_button = `<button onclick="main.VoteDelete('${objMod.link}')"> Delete </button>`;
              if (objMod.link.length > 25) {
                objMod.link = `<a href="${objMod.link}" title="${objMod.link}">${objMod.link.substring(0, 25)}\u2026</a>`;
              } else {
                objMod.link = `<a href="${objMod.link}">${objMod.link}</a>`;
              }
              if (objMod.destination.length > 25) {
                objMod.link = `<span title="${objMod.destination}">${objMod.destination.substring(0, 25)}\u2026</a>`;
              }
              dataMod.push(objMod);
            });
            resolve(dataMod);
          });
        } else {
          reject(new Error(`something went wrong [${response}]`));
        }
      })
      .catch((error) => reject(error));
  }),

  // promise of deleting a link
  voteDelete: (link, acctoken) => new Promise((resolve, reject) => {
    netRequest.voteDelete(link, acctoken)
      .then((response) => {
        if (response.status === 200) {
          resolve('voted');
        } else if (response.status === 202) {
          resolve('deleted');
        } else if (response.status === 409) {
          resolve('already voted');
        } else if (response.status === 422) {
          resolve('already deleted');
        } else if (response.status === 401) {
          reject(new Error('unauthorized'));
        } else {
          reject(new Error(`something went wrong [${response}]`));
        }
      }).catch((error) => reject(error));
  }),
};

// collection of functions to notify the user using sweetalert, all functions return a promise
const notify = {
  unauthorised: () => new Promise((resolve) => {
    Swal.fire({
      toast: true,
      position: 'top',
      showConfirmButton: false,
      title: 'Unauthoirzed',
      text: 'Please login again',
      icon: 'error',
      timer: 2000,
      timerProgressBar: true,
      didDestroy: () => {
        resolve();
      },
    });
  }),

  voteSuccess: () => new Promise((resolve) => {
    Swal.fire({
      toast: true,
      position: 'bottom-end',
      showConfirmButton: false,
      title: 'Successfully voted to delete',
      icon: 'success',
      timer: 2000,
      didDestroy: () => {
        resolve();
      },
    });
  }),

  voteDelSuccess: () => new Promise((resolve) => {
    Swal.fire({
      toast: true,
      position: 'bottom-end',
      showConfirmButton: false,
      title: 'Successfully deleted submisson.',
      text: 'Reason: 2 votes',
      icon: 'success',
      timer: 2000,
      didDestroy: () => {
        resolve();
      },
    });
  }),

  alreadyVoted: () => new Promise((resolve) => {
    Swal.fire({
      toast: true,
      position: 'top',
      showConfirmButton: false,
      title: 'You already voted to delete that link',
      icon: 'error',
      timer: 2000,
      didDestroy: () => {
        resolve();
      },
    });
  }),

  alreadyDeleted: () => new Promise((resolve) => {
    Swal.fire({
      toast: true,
      position: 'top',
      showConfirmButton: false,
      title: 'Submission already deleted',
      icon: 'info',
      timer: 2000,
      didDestroy: () => {
        resolve();
      },
    });
  }),

};

/* collection of functions to update the storage, all functions return a promise.
THEY ARE NOT SELF-CONTAINED */
const strg = {
  // promise of updating the access token
  updateAccessTkn: (refToken) => new Promise((resolve) => {
    if (!refToken) {
      notify.unauthorised().then(window.location.href = '/admin/login/');
    }

    request.newAcessTkn(refToken)
      .then((data) => {
        localStorage.setItem('accToken', data.acctoken);
        resolve();
      })
      .catch((error) => {
        if (error.message === 'unauthorized') {
          notify.unauthorised().then(window.location.href = '/admin/login/');
        } else {
          console.error(error);
        }
      });
  }),
};

const main = {

  authAndDrawTable: () => {
    $('#status').text('loading...');

    // get the access token
    strg.updateAccessTkn(localStorage.getItem('reftoken'))
      .then(() => {
        const accessToken = localStorage.getItem('accToken');
        request.getReported(accessToken)
          .then((data) => {
            $('#status').text('');
            if (data.length === 0) {
              $('#table').text('No reported links so far');
            } else {
              main.render(data);
            }
          })
          .catch((error) => {
            console.error(error);
          });
      });
  },

  render: (data) => {
    $('#table').empty();
    new gridjs.Grid({
      columns: [{
        id: 'link',
        name: 'Link',
        formatter: (_, row) => gridjs.html(`${row.cells[0].data}`),
      }, {
        id: 'destination',
        name: gridjs.html('Submitted<br>Target'),
        formatter: (_, row) => gridjs.html(`${row.cells[1].data}`),
      }, {
        id: 'times_reported',
        name: gridjs.html(' Times<br> Reported'),
      }, {
        id: 'voted_by',
        name: gridjs.html('Voted for <br>deletion by'),
        width: '30%',
      }, {
        id: 'delete_button',
        name: gridjs.html('Vote Target <br> for deletion'),
        formatter: (_, row) => gridjs.html(`${row.cells[4].data}`),
      },
      ],
      data,
      autoWidth: true,
    }).render(document.getElementById('table'));
  },

  VoteDelete: (link) => {
    const accessToken = localStorage.getItem('accToken');
    request.voteDelete(link, accessToken)
      .then((data) => {
        if (data === 'voted') {
          notify.voteSuccess();
        } else if (data === 'deleted') {
          notify.voteDelSuccess();
        } else if (data === 'already voted') {
          notify.alreadyVoted();
        } else if (data === 'already deleted') {
          notify.alreadyDeleted();
        }
      })
      .catch((error) => {
        console.error(error);
      });
  },

  parseJwt: (token) => {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map((c) => `%${(`00${c.charCodeAt(0).toString(16)}`).slice(-2)}`).join(''));

    return JSON.parse(jsonPayload);
  },
};

// Display username

$(() => {
  main.authAndDrawTable();
  // display username
  $('#username-span').text(main.parseJwt(localStorage.getItem('reftoken')).aud);

  // redraw table every 5 minutes
  setInterval(() => {
    main.authAndDrawTable();
  }, 300000);
});
