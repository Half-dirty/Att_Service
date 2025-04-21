//----GLOBAL FUNCTIONS-----------------------------------------------------------------------
//--функция для показа алерта
function showAlert(message, type = "success") {
    const alertBox = $("#custom-alert");
    alertBox.removeClass("hidden error show");
    alertBox.text(message);

    if (type === "error") {
        alertBox.addClass("error");
    }

    alertBox.addClass("show");

    // Показываем на 3 секунды, потом скрываем
    setTimeout(() => {
        alertBox.removeClass("show");
    }, 3000);
}

//---MODALS AND WORK WITH THEM-----------------------------------------
const adminModals = {
    'decline_form':
        `
    <div id="decline_form" class="popup">
        <div class="popup__body">
            <div class="popup__content">
                <div class="popup__header">
                    <a href="" class="popup__close close-popup">
                        <span></span>
                    </a>
                    <h2 class="popup__title">Укажите причину отказа:</h2>
                </div>
                <div class="popup__form">
                    <div class="popup__checker">
                        <form class="" id="decline-forma" enctype="multipart/form-data" method="POST" action="#">
                            <div class="popup__list">
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="invalid_name"> Неверно указанное ФИО
                                </label>
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="invalid_contacts"> Неверно указанные контакты
                                </label>
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="no_documents"> Прикреплены не все документы
                                </label>
                                 <div class="popup__textarea">
                                    <textarea placeholder="Напишите пояснение" name="explanation"></textarea>
                                </div>
                            </div>
                            <div class="popup__send popup__send--disabled">
                                <button type="submit" class="">Отправить</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </div>
    `,
}

function openModal(name) {
    // Добавляем модальное окно в код
    $(adminModals[name]).prependTo('.wrapper');

    const dialog = $('#' + name);

    // Делаем модальное окно видимым
    dialog.addClass('popup__open');

    $('body').addClass('modal');

    // Закрытие по крестику
    dialog.find('.close-popup').click(function (e) {
        e.preventDefault();
        closeModal(name);
    });

    // Закрытие по клавише Esc
    $(document).keydown(function (e) {
        if (e.key === 'Escape' && $('body').hasClass('modal')) {
            closeModal(name);
        }
    });

    // Закрытие по клику на фоне
    dialog.click(function (e) {
        if (!$(e.target).closest('.popup__content').length) {
            closeModal(name);
        }
    });
}

function closeModal(name) {
    const dialog = $('#' + name);
    dialog.removeClass('popup__open');
    $('body').removeClass('modal');

    // Удаляем модальное окно из DOM
    dialog.remove();
}


//---USERS PAGES-------------------------------------------------------------------

//--USER LIST PAGE-----------------------------------

//--поиск всех людей
$('#search_all_input').on('input', function (e) {
    e.preventDefault();

    let surname = $('#search_all_input').val().trim(); // Получаем введённое значение
    let content = $('.profile__user_list').empty(); // Очищаем текущий список пользователей

    // Если поле пустое, показываем весь список
    if (surname.length === 0) {
        $.ajax({
            type: "POST",
            url: "/admin/search/all",
            contentType: "application/json",
            data: JSON.stringify({ surname: '' }),  // Пустой поиск возвращает всех пользователей
            success: function (res) {
                let buf = "";
                if (res['success']) {
                    res['users'].forEach(function (user) {
                        buf += generateUserHTML(user); // Функция для генерации HTML
                    });
                } else {
                    buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
                }
                content.append(buf); // Добавляем пользователей в список
            }
        });
    } else {
        // Если поле не пустое, выполняем поиск по фамилии
        $.ajax({
            type: "POST",
            url: "/admin/search/all",
            contentType: "application/json",
            data: JSON.stringify({ surname: surname }),
            success: function (res) {
                let buf = "";
                if (res['success']) {
                    res['users'].forEach(function (user) {
                        buf += generateUserHTML(user); // Функция для генерации HTML
                    });
                } else {
                    buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
                }
                content.append(buf); // Добавляем пользователей в список
            }, error: function (xhr, status, error) {
                showAlert("Ошибка при сохранении данных!", "error");
                console.error('AJAX Error:', status, error);
            }
        });
    }
});

$(document).on('click', '.delete-student', function (e) {
    e.preventDefault();

    if (!confirm('Вы точно хотите удалить пользователя?')) {
        return;
    }

    $.ajax({
		type: "POST",
		url: "/admin/api/student",
		contentType: "application/json",
		data: JSON.stringify({ id: id }),
		success: function () {
			// Затем удаляем
			$.ajax({
				type: "POST",
				url: "/admin/student/delete",
				contentType: "application/json",
				data: JSON.stringify({}),
				success: function () {
					location.reload();
				},
				error: function () {
					showAlert("Ошибка при удалении пользователя", "error");
				}
			});
		},
		error: function () {
			showAlert("Ошибка при установке пользователя", "error");
		}
	});
});


$(document).on('click', '.open-student-profile', function (e) {
    e.preventDefault();
  
    const id = $(this).data('id');
    const source = $(this).data('source') || "";
  
    $.ajax({
      type: "POST",
      url: "/admin/api/student",
      contentType: "application/json",
      data: JSON.stringify({ id: id, source: source }),
      success: function () {
        window.location.href = "/admin/student/profile"; 
      },
      error: function () {
        showAlert("Ошибка при открытии профиля", "error");
      }
    });
  });
  


//--фильтр по ролям
$('#users_role').on('change', function (e) {
    e.preventDefault();
    let role = $('#users_role').val();
    console.log(role);
    let content = $('.profile__user_list').empty();
    $.ajax({
        type: "POST",
        url: "/admin/select",
        contentType: "application/json",
        data: JSON.stringify({ role: role }),
        success: function (res) {
            let buf = "";
            if (res.success) {
                for (let i = 0; i < res['users'].length; i++) {
                    let user = res['users'][i];
                    buf += '<div class="profile__user" id="' + user['id'] + '">' +
                        '<div class="profile__user-name">' +
                        '    <h2>' +
                        '        <img src="' + user['avatar'] + '">' +
                        '        ' + user['surname'] + ' ' + user['name'] + ' ' + user['lastname'] + '' +
                        '        <div class="header__role">';
                    switch (user['role']) {
                        case "admin":
                            buf += ' <h2 class="header__role-text header__role-text--admin">Администратор</h2>';
                            break;
                        case "student":
                            buf += '<h2 class="header__role-text header__role-text--student">Аттестуемый</h2>';
                            break;
                        case "examiner":
                            buf += '<h2 class="header__role-text header__role-text--examiner">Экзаменатор</h2>';
                            break;
                    }
                    buf += '</div>' +
                        '   </h2>' +
                        '</div>' +
                        '<div class="profile__user-selector">' +
                        '    <ul class="profile__user-menu">' +
                        '        <li><button class="profile__user-link open-student-profile" data-id="{{.ID}}">Посмотреть аккаунт</button></li>' +
                        '        <li><button class="profile__user-link delete-student" data-id="{{.ID}}">Удалить аккаунт</button></li>' +
                        '    </ul>' +
                        '</div>' +
                        '</div>'
                }
            } else {
                buf += '<div class="profile__user">' +
                    '    <div class="profile__user-name">' +
                    '       <h2>пользователи не найдены</h2>' +
                    '    </div>' +
                    '   <div class="profile__user-selector">' +
                    '        <div class="profile__menu-icon">' +
                    '            <span></span>' +
                    '        </div>' +
                    '       </div>' +
                    '</div>'
            }
            content.append(buf);
        }
    })
})


//--USER APPLICATIONS PAGE-----------------------------------

// Слушаем ввод в поле для поиска в реальном времени
$('#search_application_input').on('input', function (e) {
    e.preventDefault();

    let surname = $('#search_application_input').val().trim(); // Получаем введённое значение
    let content = $('.profile__user_list').empty(); // Очищаем текущий список пользователей

    // Если поле пустое, показываем весь список
    if (surname.length === 0) {
        $.ajax({
            type: "POST",
            url: "/admin/search/application",
            contentType: "application/json",
            data: JSON.stringify({ surname: '' }),  // Пустой поиск возвращает всех пользователей
            success: function (res) {
                let buf = "";
                if (res['success']) {
                    res['users'].forEach(function (user) {
                        buf += generateUserHTML(user); // Функция для генерации HTML
                    });
                } else {
                    buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
                }
                content.append(buf); // Добавляем пользователей в список
            }
        });
    } else {
        // Если поле не пустое, выполняем поиск по фамилии
        $.ajax({
            type: "POST",
            url: "/admin/search/application",
            contentType: "application/json",
            data: JSON.stringify({ surname: surname }),
            success: function (res) {
                let buf = "";
                if (res['success']) {
                    res['users'].forEach(function (user) {
                        buf += generateUserHTML(user); // Функция для генерации HTML
                    });
                } else {
                    buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
                }
                content.append(buf); // Добавляем пользователей в список
            }, error: function (xhr, status, error) {
                showAlert("Ошибка при сохранении данных!", "error");
                console.error('AJAX Error:', status, error);
            }
        });
    }
});

// Функция для генерации HTML для каждого пользователя
function generateUserHTML(user) {
    let buf = '<div class="profile__user" id="' + user['id'] + '">' +
        '<div class="profile__user-name">' +
        '    <h2>' +
        '        <img src="' + user['avatar'] + '">' +
        '        ' + user['surname'] + ' ' + user['name'] + ' ' + user['lastname'] + '' +
        '        <div class="header__role">';
    switch (user['role']) {
        case "admin":
            buf += ' <h2 class="header__role-text header__role-text--admin">Администратор</h2>';
            break;
        case "student":
            buf += '<h2 class="header__role-text header__role-text--student">Аттестуемый</h2>';
            break;
        case "examiner":
            buf += '<h2 class="header__role-text header__role-text--examiner">Экзаменатор</h2>';
            break;
    }
    buf += '</div>' +
        '   </h2>' +
        '</div>' +
        '<div class="profile__user-selector">' +
        '    <ul class="profile__user-menu">' +
        '        <li><button class="profile__user-link open-student-profile" data-id="{{.ID}}" data-source="application">Посмотреть аккаунт</button></li>' +
        '        <li><button class="profile__user-link delete-student" data-id="{{.ID}}">Удалить аккаунт</button></li>' +
        '    </ul>' +
        '</div>' +
        '</div>';
    return buf;
}

// Функция для вывода, если пользователей не найдено
function generateUserNotFound() {
    return '<div class="profile__user">' +
        '    <div class="profile__user-name">' +
        '       <h2>Пользователи не найдены</h2>' +
        '    </div>' +
        '   <div class="profile__user-selector">' +
        '        <div class="profile__menu-icon">' +
        '            <span></span>' +
        '        </div>' +
        '       </div>' +
        '</div>';
}


//--фильтр по ролям (хз зачем)
$('#users_role_application').on('change', function (e) {
    e.preventDefault();
    let role = $('#users_role_application').val();
    console.log(role);
    let content = $('.profile__user_list').empty();
    $.ajax({
        type: "POST",
        url: "/admin/select/application",
        contentType: "application/json",
        data: JSON.stringify({ role: role }),
        success: function (res) {
            let buf = "";
            if (res.success) {
                for (let i = 0; i < res['users'].length; i++) {
                    let user = res['users'][i];
                    buf += '<div class="profile__user" id="' + user['id'] + '">' +
                        '<div class="profile__user-name">' +
                        '    <h2>' +
                        '        <img src="' + user['avatar'] + '">' +
                        '        ' + user['surname'] + ' ' + user['name'] + ' ' + user['lastname'] + '' +
                        '        <div class="header__role">';
                    switch (user['role']) {
                        case "admin":
                            buf += ' <h2 class="header__role-text header__role-text--admin">Администратор</h2>';
                            break;
                        case "student":
                            buf += '<h2 class="header__role-text header__role-text--student">Аттестуемый</h2>';
                            break;
                        case "examiner":
                            buf += '<h2 class="header__role-text header__role-text--examiner">Экзаменатор</h2>';
                            break;
                    }
                    buf += '</div>' +
                        '   </h2>' +
                        '</div>' +
                        '<div class="profile__user-selector">' +
                        '    <ul class="profile__user-menu">' +
                        '        <li><button class="profile__user-link open-student-profile" data-id="{{.ID}}" data-source="application">Посмотреть аккаунт</button></li>' +
                        '        <li><button class="profile__user-link delete-student" data-id="{{.ID}}">Удалить аккаунт</button></li>' +
                        '    </ul>' +
                        '</div>' +
                        '</div>'
                }
            } else {
                buf += '<div class="profile__user">' +
                    '    <div class="profile__user-name">' +
                    '       <h2>пользователи не найдены</h2>' +
                    '    </div>' +
                    '   <div class="profile__user-selector">' +
                    '        <div class="profile__menu-icon">' +
                    '            <span></span>' +
                    '        </div>' +
                    '       </div>' +
                    '</div>'
            }
            content.append(buf);
        }
    })
})


//--USER SHOW PAGES-----------------------------------

//--изменение роли пользователя (выпадающий список)
$('#role-select').on('change', function (e) {
    e.preventDefault();

    let role = $(this).val();

    // Обновляем класс для визуального оформления
    $(this)
        .removeClass('header__role-select--student header__role-select--examiner')
        .addClass(`header__role-select--${role}`);

    // Безопасный AJAX-запрос без использования ID в URL
    $.ajax({
        type: "POST",
        url: "/admin/change_role", // безопасный путь
        contentType: "application/json",
        data: JSON.stringify({ role: role }),
        success: function (res) {
            if (res.success) {
                showAlert("Роль успешно обновлена!");
            } else {
                showAlert("Ошибка при обновлении роли!", "error");
            }
        },
        error: function (xhr, status, error) {
            showAlert("Ошибка при обновлении роли!", "error");
            console.error('AJAX Error:', status, error);
        }
    });
});




//--подтвердить профиль
$('#decision-accept').on('click', function (e) {
	e.preventDefault();
	$.ajax({
		type: "POST",
		url: "/admin/student/confirm",
		contentType: "application/json",
		data: JSON.stringify({ confirm: true }),
		success: function (res) {
			if (res.success) {
				window.location.href = "/admin/user/application";
			} else {
				showAlert("Не удалось подтвердить пользователя", "error");
			}
		}
	});
});


//--отклонить профиль (открытие модалки)
$('.profile__decision-decline').on('click', function (e) {
    e.preventDefault();
    openModal("decline_form");
})

//--функция, для активации кнопки на отправку формы
$(document).on('change keyup', '#decline_form input[type="checkbox"], #decline_form textarea', function () {
    const isChecked = $('#decline_form input[type="checkbox"]:checked').length > 0;
    const isTextEntered = $('#decline_form textarea').val().trim().length > 0;

    if (isChecked || isTextEntered) {
        $('.popup__send').removeClass('popup__send--disabled');
    } else {
        $('.popup__send').addClass('popup__send--disabled');
    }
});


//--отправка формы отказа
$(document).on('submit', '#decline-forma', function (e) {
    e.preventDefault();

    if ($('.popup__send').hasClass('popup__send--disabled')) {
        return; // Не отправляем запрос, если кнопка не активна
    }
    ;
    const formData = {};

    // Собираем все выбранные чекбоксы
    $('input[name="reason"]:checked').each(function () {
        if (!formData['reasons']) formData['reasons'] = [];
        formData['reasons'].push($(this).val());
    });

    // Добавляем текстовое пояснение, если оно есть
    const explanation = $('textarea[name="explanation"]').val();
    if (explanation !== undefined && explanation.trim() !== '') {
        formData['explanation'] = explanation;
    }

    $.ajax({
        type: "POST",
        url: "/admin/student/decline", // безопасный путь
        contentType: "application/json",
        data: JSON.stringify({ reasons: reasons, explanation: explanation }),
        success: function () {
            window.location.href = "/admin/user/application";
        }
    });    
});



//----EXAM PAGES
//--EXAM CREATE PAGE-----------------------------------
//--при загрузке страницы, фокус на поле даты экзамена
$(document).ready(function () {
    $('#exam_date').focus();
});

//---Examiner list--------------------
$('#exam_code').inputmask('99-99-99', { autoUnmask: true });

function setupList({ selectorIcon, selectorButton, labelOn, labelOff, roleSelectorClass }) {
    const icons = document.querySelectorAll(selectorIcon);
    const button = document.querySelector(selectorButton);

    // вспомогательная функция проверки "все выбраны?"
    function checkAllSelected() {
        return Array.from(icons).every(icon => icon.classList.contains('active'));
    }

    // обновление текста кнопки в зависимости от состояния
    function updateButtonLabel() {
        const allSelected = checkAllSelected();
        button.textContent = allSelected ? labelOff : labelOn;
    }

    // обработчик кликов по иконке
    icons.forEach(icon => {
        icon.addEventListener('click', () => {
            icon.classList.toggle('active');

            const roleSelector = icon.closest('.profile__examiner-item')?.querySelector(roleSelectorClass);
            if (icon.classList.contains('active')) {
                roleSelector?.style.setProperty('display', 'block');
            } else {
                roleSelector?.style.setProperty('display', 'none');
            }

            updateButtonLabel(); // обновляем текст кнопки при каждом клике
        });
    });

    // обработчик кнопки "выбрать/убрать всех"
    button.addEventListener('click', function (e) {
        e.preventDefault();
        const allSelected = checkAllSelected();
        const shouldActivate = !allSelected;

        icons.forEach(icon => {
            icon.classList.toggle('active', shouldActivate);

            const roleSelector = icon.closest('.profile__examiner-item')?.querySelector(roleSelectorClass);
            if (shouldActivate) {
                roleSelector?.style.setProperty('display', 'block');
            } else {
                roleSelector?.style.setProperty('display', 'none');
            }
        });

        updateButtonLabel();
    });
}



$(document).on('click', '.profile__examiner-item .profile__menu-icon', function (e) {
    e.preventDefault();
    let iconStatus = $(this).hasClass('active');
    const roleSelect = $(this).closest('.profile__examiner-item').find('.profile__examiner-role');
    if (iconStatus) {
        roleSelect.show();
    } else {
        roleSelect.hide();
    }
});

setupList({
    selectorIcon: '.profile__examiner-item .profile__menu-icon',
    selectorButton: '#select-all_examiner',
    labelOn: 'Выбрать всех',
    labelOff: 'Убрать всех',
    roleSelectorClass: '.profile__examiner-role'
});

setupList({
    selectorIcon: '.profile__student-item .profile__menu-icon',
    selectorButton: '#select-all_student',
    labelOn: 'Выбрать всех',
    labelOff: 'Убрать всех',
    roleSelectorClass: ''
});


function getUserList(role) {
    const roles = {
        'examiner': {
            icon: '.profile__examiner-item .profile__menu-icon',
            item: '.profile__examiner-item',
            roleSelectorClass: '.profile__examiner-role'
        },
        'student': {
            icon: '.profile__student-item .profile__menu-icon',
            item: '.profile__student-item',
            roleSelectorClass: ''
        }
    }

    const icons = document.querySelectorAll(roles[role].icon);
    let users = [];
    icons.forEach(icon => {
        if (icon.classList.contains('active')) {
            const item = icon.closest(roles[role].item);
            if (item) {
                users.push(item.getAttribute('data-id'));
                if (role === 'examiner') {
                    const roleSelect = item.querySelector(roles[role].roleSelectorClass);
                    if (roleSelect) {
                        users.push(roleSelect.value); // добавляем роль экзаменатора (не уверен, что работает)
                    }
                }
            }
        }
    });

    return users;
}

//--Отправка формы создания экзамена
function submitExamForm(link) {
    let formData = new FormData();
    let examiners = [];
    let students = [];
    let date = $('#exam_date').val();
    let commissionStart = $('#commission_start').val();
    let commissionEnd = $('#commission_end').val();

    // Получаем всех выбранных экзаменаторов
    examiners = getUserList('examiner');
    // Получаем всех выбранных студентов
    students = getUserList('student');

    // Добавляем экзаменаторов в FormData
    formData.append('examiners', JSON.stringify(examiners));
    formData.append('students', JSON.stringify(students));
    formData.append('date', date);
    formData.append('commission_start', commissionStart);
    formData.append('commission_end', commissionEnd);

    $.ajax({
        type: "POST",
        url: link,
        cache: false,
        processData: false,
        contentType: false,
        data: formData,
        success: function (res) {
            if (res.success) {
                showAlert("Данные успешно сохранены!");
                return window.location.href = "/admin/exam/planning";
            } else {
                showAlert("Ошибка при сохранении данных!", "error");
            }
        }, error: function (xhr, status, error) {
            showAlert("Ошибка при сохранении данных!", "error");
            console.error('AJAX Error:', status, error);
        }
    });
};

// Отправка по кнопке
$('#create_exam').on('click', function (e) {
    e.preventDefault();
    submitExamForm("/admin/exam/create");
});

$('#assign_exam').on('click', function (e) {
    e.preventDefault("/admin/exam/assign");
    
});

// Отправка по Enter внутри любых input'ов
$(document).ready(function () {
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

//--Exam Application Page-----------------------------------
$('#agree_application').on('click', function (e) {
    e.preventDefault();

    //-----------------------------------------------------------------------
})