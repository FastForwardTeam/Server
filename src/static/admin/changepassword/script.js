domain = 'https://crowd.fastforward.team/'

$('#changepass').on("click", function(){
$('#changepass').attr('disabled',true);
let u = $('#username').val()
let o = $('#oldpassword').val()
let n = $('#newpassword').val()
fetch(domain + 'admin/api/changepassword', {
method: "post",
headers: {
  'Accept': 'application/json',
  'Content-Type': 'application/json'
},
body: JSON.stringify({
  username: u,
  oldpassword: o,
  newpassword: n
   })
})
  .then(res => {
    if (res.status == 200) {
      Swal.fire({
        toast: true,
        position: 'top',
        showConfirmButton: false,
        title: 'Password changed successfully, please login with your new password.',
        icon: 'success',
        timer: 2000,
        timerProgressBar: true,
        didDestroy: () => {
          window.location.replace("../login");
        }
      })
    } else {
      $('#changepass').attr('disabled',false);
      Swal.fire({
        toast: true,
        position: 'top',
        showConfirmButton: false,
        title: 'Incorrect username or password',
        icon: 'error',
        timer: 2000,
        timerProgressBar: true,
      })
    }
    })
})
