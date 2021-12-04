domain = 'https://crowd.fastforward.team/'

$('#login').on("click", function(){
$('#login').attr('disabled',true);
let u = $('#username').val()
let p = $('#password').val()
fetch(domain+'admin/api/newreftoken', {
method: "post",
headers: {
  'Accept': 'application/json',
  'Content-Type': 'application/json'
},
body: JSON.stringify({
  username: u,
  password: p
   })
})
  .then(res => res.json())
  .then(data => {
    localStorage.setItem('reftoken', data.reftoken);
    window.location.replace("../");
  })
  .catch(error => {
    $('#login').attr('disabled',false);
    Swal.fire({
      toast: true,
      position: 'top',
      showConfirmButton: false,
      title: 'Incorrect username or password',
      icon: 'error',
      timer: 2000,
      timerProgressBar: true,
    })
  })
})
