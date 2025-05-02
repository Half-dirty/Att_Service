//--при загрузке страницы, фокус на поле почты
$(document).ready(function () {
    $('#email').focus();
});

//--при загрузке страницы, фокус на поле пароля
$(document).ready(function () {
    $('#password').focus();
});


$(document).ready(function (e) {
    $(document).on('keydown', 'button', function () {
        if (e.key === 'Enter') {
            e.preventDefault();
            $('#send__main_page').click();
        }
    });
    $(document).on('keydown', 'input, select, textarea', function (e) {
        if (e.key === 'Enter') {
            e.preventDefault(); // если не хочешь чтобы он случайно форму "отправил" куда-то

            const inputs = $('input, select, textarea')
                .filter(':visible:not([disabled])');

            const idx = inputs.index(this);
            if (idx > -1 && idx + 1 < inputs.length) {
                inputs.eq(idx + 1).focus();
            }
        }
    });
});



/* 
LOGIN FORM
    get -
    post {
        email: email
        password: password
    }

    after
        get {
            success: true/false
            error: type
            link: link to student/admin page
        }

    отправка почты и пароля
    после отправки ждет ответ сервера
    в ответе ожидает подтверждение или отказа - true/false
    
    если success - true
    то смотрит ссылку, которую присылает сервер. Ссылка должна быть на страницу соответствующую
    т.е. если залогинился студент - то ссылка student, иначе - admin
    даже если пользователь - член комиссии, он идет на /student

    если success - false
    то смотрит причину - error
    в ошибке можешь указать текстом, например если неверный пароль - wrong_pass
    если неверное имя - wrong_name

    ЕСЛИ НЕ ЗАХОЧЕШЬ НАД ЭТИМ ПАРИТСЯ СООБЩИ
    в таком случае error не нужен

    при success - true клиент сам перенаправит на соответствующую страницу
*/
console.log("JS загружен");

$('#login-button').on('click', function (e) {
    e.preventDefault();
    let email = $('#email').val();
    let pass = $('#password').val();
    $.ajax({
        type: "POST",
        url: "/login",
        contentType: "application/json",
        data: JSON.stringify({ email: email, pass: pass }),
        success: function (res) {
            if (res.success) {
                localStorage.setItem("refresh_token", res.refresh_token); // ← обязательно
                return window.location.href = `/${res.link}`;
            }
        
            if (res.error) {
                switch (res.error) {
                    case "userNone":
                        $("#email").addClass("not-correct");
                        break;
                    case "passwordNone":
                        $("#pass").addClass("not-correct");
                        break;
                }
            }
        },
        error: function (xhr, status, error) {
            console.error('AJAX Error:', status, error);
        }
    })
})




/* 
REGISTRATION FORM
    get -
    post {
        email: email
    }

    after
        get -

    отправка почты
    после отправки тупо перекидывает на confirm (подтверждение)
*/
$('#registration-button').on('click', function (e) {
    e.preventDefault();
    let email = $('#email').val();
    $.ajax({
        type: "POST",
        url: "/registration",
        contentType: "application/json",
        data: JSON.stringify({ email: email }),
        success: function (res) {
            return window.location.href = `/`;
        },
        error: function (xhr, status, error) {
            console.error('AJAX Error:', status, error);
        }
    })
})


const refreshAccessToken = () => {
    $.ajax({
        type: 'POST',
        url: '/refresh',
        xhrFields: { withCredentials: true }, // вот это очень важно!!
        success: function (res) {
            if (res.success === true) {
                console.log("Токен успешно обновлен");
            } else {
                console.warn("Ошибка при обновлении токена:", res);
                window.location.href = "/";
            }
        }
    });
};


$(document).ready(function () {
    setTimeout(refreshAccessToken, 1000); // через 1 секунду
    setInterval(refreshAccessToken, 10 * 60 * 1000);
});





$('#confirm-button').on('click', function (e) {
    e.preventDefault();
    let email = $('#email').val();
    let password = $('#password').val();
    $.ajax({
        type: "POST",
        url: "/confirm",
        contentType: "application/json",
        data: JSON.stringify({ email: email, password: password }),
        success: function (res) {
            return window.location.href = `/`;
        },
        error: function (xhr, status, error) {
            console.error('AJAX Error:', status, error);
        }
    })
})
