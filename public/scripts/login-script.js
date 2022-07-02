const form = document.getElementById('login-form')

form.addEventListener('submit', (event) => {
    event.preventDefault()
    
    const usrname = document.getElementById("usrname").value;
    const passwd = document.getElementById("passwd").value;
    const payload = "username=".concat(usrname).concat("&password=").concat(passwd);
    const paramHeaders = new Headers({'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'});

    var token = ""
    fetch("/request-access", {
        method: 'POST',
        body: payload,
        headers: paramHeaders}
         ).then(res => res.json())
        .then(res => {
            if(typeof res.token == 'undefined') {
                window.location.reload()
            } else {
                token = res.token
                console.log(res.token)
            }
        })
})
