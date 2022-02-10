$('#login').on('click', () => {
  $('#login').attr('disabled', true);
  const u = $('#username').val();
  const p = $('#password').val();
  fetch('/admin/api/newreftoken', {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      username: u,
      password: p,
    }),
  })
    .then((res) => res.json())
    .then((data) => {
      localStorage.setItem('reftoken', data.reftoken);
      window.location.replace('/admin/');
    })
    .catch((error) => {
      $('#login').attr('disabled', false);
      Swal.fire({
        toast: true,
        position: 'top',
        showConfirmButton: false,
        title: 'Incorrect username or password',
        icon: 'error',
        timer: 2000,
        timerProgressBar: true,
      });
    });
});
