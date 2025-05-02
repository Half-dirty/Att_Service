
//----GLOBAL FUNCTIONS------------------------------------------------------

const apiPost = (url, data) => {
    return $.ajax({
        type: 'POST',
        url,
        contentType: 'application/json',
        data: JSON.stringify(data),
    });
};


const showAlert = (message, type = 'success') => {
    const alertBox = $("#custom-alert").removeClass("hidden error show").text(message);
    alertBox.addClass(type === "error" ? "error" : "").addClass("show");
    setTimeout(() => alertBox.removeClass("show"), 3000);
};

const generateUserNotFound = () => `
<div class="profile__user">
    <div class="profile__user-name">
        <h2>Пользователи не найдены</h2>
    </div>
</div>`;


const adminModals = {
    decline_form: `
    <div id="decline_form" class="popup">
        <div class="popup__body">
            <div class="popup__content">
                <div class="popup__header">
                    <a href="#" class="popup__close close-popup"><span></span></a>
                    <h2 class="popup__title">Укажите причину отказа:</h2>
                </div>
                <form id="decline-forma">
                    <label><input type="checkbox" name="reason" value="invalid_name"> Неверно указанное ФИО</label>
                    <label><input type="checkbox" name="reason" value="invalid_contacts"> Неверно указанные контакты</label>
                    <label><input type="checkbox" name="reason" value="no_documents"> Не все документы</label>
                    <textarea name="explanation" placeholder="Пояснение"></textarea>
                    <button type="submit">Отправить</button>
                </form>
            </div>
        </div>
    </div>`,
    decline_application: `
<div id="decline_application_modal" class="popup">
    <div class="popup__body">
        <div class="popup__content">
            <div class="popup__header">
                <a href="#" class="popup__close close-popup"><span></span></a>
                <h2 class="popup__title">Укажите причину отказа:</h2>
            </div>
            <form id="decline_application-form" class="popup__form">
                <div class="popup__checker">
                    <div class="popup__list">
                        <label class="popup__item">
                            <input type="checkbox" name="reason" value="invalid_name"> Неверно указанные данные
                        </label>
                        <label class="popup__item">
                            <input type="checkbox" name="reason" value="invalid_contacts"> Прикреплены не соответствующие фото
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
                </div>
            </form>
        </div>
    </div>
</div>
`,

    'start-exam': `
<div id="start-exam" class="popup">
    <div class="popup__body">
        <div class="popup__content">
            <div class="popup__header">
                <a href="#" class="popup__close close-popup"><span></span></a>
                <h2 class="popup__title">Экзаменаторы подключившиеся к экзамену:</h2>
            </div>
            <div class="popup__form">
                <div class="popup__examiner-list" id="examiner_list">
                    <!-- сюда динамически вставим экзаменаторов -->
                </div>
                <div class="popup__start">
                    <button type="button" id="start_exam_confirm">Начать экзамен</button>
                </div>
            </div>
        </div>
    </div>
</div>`,

};

//---MODAL CLASS------------------------------------------------------------

class Modal {
    constructor(id, template) {
        this.id = id;
        this.template = template;
    }

    open() {
        $(this.template).prependTo('.wrapper').addClass('popup__open');
        $('body').addClass('modal');
        this.bindEvents();
    }

    close() {
        $(`#${this.id}`).removeClass('popup__open').remove();
        $('body').removeClass('modal');
    }

    bindEvents() {
        const self = this;
        $(document).on('click', '.popup__close, .popup__body', function (e) {
            if (!$(e.target).closest('.popup__content').length) {
                e.preventDefault();
                self.close();
            }
        });

        $(document).keydown(function (e) {
            if (e.key === 'Escape') self.close();
        });
    }
}

const declineModal = new Modal('decline_form', adminModals.decline_form);

//---TEMPLATES--------------------------------------------------------------

const userCard = ({ id, avatar, surname, name, lastname, role }) => `
<div class="profile__user" data-id="${id}">
    <div class="profile__user-name">
        <h2>
            <img src="${avatar}" />
            ${surname} ${name} ${lastname}
            <div class="header__role">${role}</div>
        </h2>
    </div>
    <ul class="profile__user-menu">
        <li><button class="profile__user-link open-student-profile" data-id="${id}">Посмотреть</button></li>
        <li><button class="profile__user-link delete-student" data-id="${id}">Удалить аккаунт</button></li>
    </ul>
</div>`;


//---EVENT HANDLERS---------------------------------------------------------

$('.profile__user_list').on('click', '.delete-student', function () {
    if (!confirm('Вы точно хотите удалить пользователя?')) return;

    const id = $(this).data('id');
    apiPost('/admin/api/student', { id })
        .done(() => apiPost('/admin/student/delete', {}).done(() => location.reload()))
        .fail(() => showAlert('Ошибка при удалении пользователя', 'error'));
});

$('.profile__user_list').on('click', '.open-student-profile', function () {
    const id = $(this).data('id');
    const source = $(this).data('source') || "";
    apiPost('/admin/api/student', { id, source })
        .done(() => window.location.href = '/admin/student/profile')
        .fail(() => showAlert('Ошибка при открытии профиля', 'error'));
});

$('#decision-accept').on('click', function (e) {
    e.preventDefault();
    apiPost('/admin/student/confirm', { confirm: true })
        .done(() => window.location.href = "/admin/user/application")
        .fail(() => showAlert("Ошибка подтверждения пользователя", "error"));
});

$('#role-select').change(function () {
    const role = $(this).val();
    $(this).toggleClass('header__role-select--examiner', role === 'examiner').toggleClass('header__role-select--student', role !== 'examiner');

    apiPost('/admin/change_role', { role })
        .done(() => showAlert('Роль успешно обновлена!'))
        .fail(() => showAlert('Ошибка при обновлении роли!', 'error'));
});

//---TOKEN REFRESH FIX------------------------------------------------------

// этот код нужен, чтобы обновлять access_token "в фоне"
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


//---NAVIGATION FIX---------------------------------------------------------

// $('nav.menu').on('click', 'a.menu__button', function (e) {
//     e.preventDefault();
//     const link = $(this).attr('href');

//     $.ajax({
//         url: link,
//         headers: { 'Accept': 'text/html' },
//         success: () => window.location.href = link,
//         error: (xhr) => xhr.status === 401 ? window.location.href = '/' : showAlert('Ошибка доступа к странице', 'error')
//     });
// });

$('nav.menu').on('click', 'a.menu__button', function (e) {
    e.preventDefault();
    window.location.href = $(this).attr('href');
});

//---INIT PAGE EVENTS-------------------------------------------------------
$(document).ready(() => {
    $('#exam_date').focus();
    updateExamButtonState(); // 🔥 Добавляем здесь — при загрузке страницы проверяет студентов
});
//---ПОИСК ПОЛЬЗОВАТЕЛЕЙ ПО ФАМИЛИИ----------------------------------------

$('#search_all_input, #search_application_input').on('input', function () {
    const surname = $(this).val().trim();
    const content = $('.profile__user_list').empty();

    apiPost('/admin/search/all', { surname })
        .done(res => {
            if (res.success && res.users.length) {
                res.users.forEach(user => content.append(userCard(user)));
            } else {
                content.append('<div>Пользователи не найдены</div>');
            }
        })
        .fail(() => showAlert('Ошибка поиска пользователей', 'error'));
});

//---ФИЛЬТР ПО РОЛЯМ-------------------------------------------------------

$('#users_role, #users_role_application').on('change', function () {
    const role = $(this).val();
    const url = $(this).attr('id') === 'users_role' ? '/admin/select' : '/admin/select/application';
    const content = $('.profile__user_list').empty();

    apiPost(url, { role })
        .done(res => {
            if (res.success && res.users.length) {
                res.users.forEach(user => content.append(userCard(user)));
            } else {
                content.append('<div>Пользователи не найдены</div>');
            }
        })
        .fail(() => showAlert('Ошибка фильтрации по роли', 'error'));
});

//---ОТМЕНА ЭКЗАМЕНОВ------------------------------------------------------

$(document).on("click", ".profile__exam-link--cancel", function (e) {
    e.preventDefault();
    const examId = $(this).closest(".profile__exam").data("id");

    if (!examId) return alert("ID экзамена не найден");

    apiPost("/admin/api/exam/cancel", { exam_id: examId })
        .done(() => location.reload())
        .fail(xhr => alert("Ошибка отмены экзамена: " + xhr.responseText));
});

//---ОТКРЫТИЕ МОДАЛКИ ОТКЛОНЕНИЯ ПРОФИЛЯ-----------------------------------

$(document).on('click', '.profile__decision-decline', function (e) {
    e.preventDefault();
    declineModal.open();
});

//---АКТИВАЦИЯ КНОПКИ ОТПРАВКИ ФОРМЫ ОТКЛОНЕНИЯ-----------------------------

$(document).on('change keyup', '#decline_form input[type="checkbox"], #decline_form textarea', function () {
    const active = $('#decline_form input[type="checkbox"]:checked').length > 0 || $('#decline_form textarea').val().trim().length > 0;
    $('.popup__send').toggleClass('popup__send--disabled', !active);
});

//---ОТПРАВКА ФОРМЫ ОТКАЗА--------------------------------------------------

$(document).on('submit', '#decline-forma', function (e) {
    e.preventDefault();
    if ($('.popup__send').hasClass('popup__send--disabled')) return;

    const reasons = $('#decline_form input[name="reason"]:checked').map((_, el) => el.value).get();
    const explanation = $('#decline_form textarea[name="explanation"]').val();
    const appID = $('body').data('id');

    apiPost("/admin/student/decline", { id: appID, reasons, explanation })
        .done(() => window.location.href = "/admin/user/application")
        .fail(xhr => alert("Ошибка при отправке: " + xhr.responseText));
});

//---ФУНКЦИЯ ВЫБОРА ВСЕХ ЭКЗАМЕНАТОРОВ И СТУДЕНТОВ--------------------------
function setupList({ selectorIcon, selectorButton, labelOn, labelOff, roleSelectorClass }) {
    const icons = $(selectorIcon);
    const button = $(selectorButton);

    button.on('click', function (e) {
        e.preventDefault();
        const allSelected = icons.length > 0 && icons.length === icons.filter('.active').length;
        const newState = !allSelected;

        icons.toggleClass('active', newState);

        if (roleSelectorClass) {
            icons.each(function () {
                const $roleSelect = $(this).closest('.profile__examiner-item').find(roleSelectorClass);
                if (newState) {
                    $roleSelect.show();
                } else {
                    $roleSelect.val('examiner'); // ❗ При убирании — сбрасываем роль на "examiner"
                    $roleSelect.hide();
                }
            });
        }

        updateButtonLabel();
        updateExamButtonState();
    });

    updateButtonLabel();
}




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

$('#select-all_student').on('click', function () {
    setTimeout(updateExamButtonState, 50); // даём время DOM обновиться после массового выбора
});


//---ОТПРАВКА ФОРМЫ СОЗДАНИЯ ЭКЗАМЕНА--------------------------------------// Нормализация даты перед отправкой
function normalizeDate(dateStr) {
    if (!dateStr) return "";
    if (dateStr.includes(".")) {
        const [day, month, year] = dateStr.split(".");
        return `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
    }
    return dateStr; // если уже yyyy-mm-dd
}

// Обновление состояния кнопок Сохранить / Назначить
function updateExamButtonState() {
    const studentsSelected = $('.profile__student-item .profile__menu-icon.active').length > 0;
    $('#create_exam, #assign_exam').prop('disabled', !studentsSelected);
}

// Сбор выбранных пользователей
function getUserList(role) {
    const roles = {
        examiner: '.profile__people-examiner .profile__examiner-item .profile__menu-icon.active',
        student: '.profile__people-student .profile__student-item .profile__menu-icon.active'
    };

    let users = [];
    $(roles[role]).each(function () {
        users.push($(this).closest('[data-id]').data('id'));
    });
    return users;
}

// Отправка формы экзамена
function submitExamForm(link, autoSchedule) {
    const examiners = getUserList('examiner');
    const students = getUserList('student');
    const date = $('#exam_date').val();
    const commissionStart = $('#commission_start').val();
    const commissionEnd = $('#commission_end').val();
    let chairmanId = null, secretaryId = null;

    $('.profile__people-examiner .profile__examiner-item .profile__menu-icon.active').each(function () {
        const role = $(this).closest('.profile__examiner-item').find('.profile__examiner-role').val();
        const id = $(this).closest('.profile__examiner-item').data('id');
        if (role === 'chair') chairmanId = id;
        if (role === 'secretary') secretaryId = id;
    });

    if (!date || !commissionStart || !commissionEnd) {
        showAlert("Заполните все даты экзамена!", "error");
        return;
    }
    if (students.length === 0) {
        showAlert("Выберите хотя бы одного экзаменуемого!", "error");
        return;
    }

    const formData = new FormData();
    formData.append('examiners', JSON.stringify(examiners));
    formData.append('students', JSON.stringify(students));
    formData.append('date', normalizeDate(date));
    formData.append('commission_start', normalizeDate(commissionStart));
    formData.append('commission_end', normalizeDate(commissionEnd));
    formData.append('auto_schedule', autoSchedule ? 'true' : 'false');
    if (chairmanId !== null) formData.append('chairman_id', JSON.stringify(chairmanId));
    if (secretaryId !== null) formData.append('secretary_id', JSON.stringify(secretaryId));

    $.ajax({
        type: "POST",
        url: link,
        processData: false,
        contentType: false,
        data: formData,
        success: (res) => {
            if (res.success) {
                showAlert("Экзамен успешно сохранён!");
                window.location.href = autoSchedule ? "/admin/exam/scheduled" : "/admin/exam/planning";
            } else {
                showAlert("Ошибка при сохранении!", "error");
            }
        },
        error: () => showAlert("Ошибка при сохранении!", "error")
    });
}

// Навешивание событий на иконки через делегирование
$(document).on('click', '.profile__examiner-item .profile__menu-icon', function () {
    $(this).toggleClass('active');
    $(this).closest('.profile__examiner-item').find('.profile__examiner-role').toggle($(this).hasClass('active'));
    updateExamButtonState();
});

// Делегированный обработчик клика на экзаменаторов и студентов
$(document).on('click', '.profile__examiner-item, .profile__student-item', function (e) {
    // Если клик был по select внутри экзаменатора — ничего не делать
    if ($(e.target).is('select')) {
        e.stopPropagation(); // Остановить всплытие!
        return;
    }

    const $item = $(this);
    const $icon = $item.find('.profile__menu-icon');

    $icon.toggleClass('active');

    if ($item.hasClass('profile__examiner-item')) {
        const $roleSelect = $item.find('.profile__examiner-role');
        if ($icon.hasClass('active')) {
            $roleSelect.show();
        } else {
            $roleSelect.val('examiner'); // ❗ При снятии выбора сбрасываем на экзаменатора
            $roleSelect.hide();
        }
    }

    updateButtonLabel();
    updateExamButtonState();
});

function updateButtonLabel() {
    const examinerIcons = $('.profile__examiner-item .profile__menu-icon');
    const studentIcons = $('.profile__student-item .profile__menu-icon');

    const allExaminersSelected = examinerIcons.length > 0 && examinerIcons.length === examinerIcons.filter('.active').length;
    const allStudentsSelected = studentIcons.length > 0 && studentIcons.length === studentIcons.filter('.active').length;

    $('#select-all_examiner').text(allExaminersSelected ? 'Убрать всех' : 'Выбрать всех');
    $('#select-all_student').text(allStudentsSelected ? 'Убрать всех' : 'Выбрать всех');
}
function openModal(name) {
    const modalTemplate = adminModals[name];  // Получаем шаблон модалки
    if (!modalTemplate) {
        console.error("Модалка не найдена:", name);
        return;
    }

    // Удаляем предыдущую модалку, если такая уже существует
    $('.popup#' + name).remove(); 
    // Добавляем новую модалку в контейнер
    $('.wrapper').append(modalTemplate);  

    // Добавляем классы для отображения модалки
    $('#' + name).addClass('popup__open');
    $('body').addClass('modal');

    // Закрытие модалки по крестику
    $('#' + name).find('.popup__close').on('click', function (e) {
        e.preventDefault();
        closeModal(name);
    });

    // Закрытие по клику вне модалки
    $(document).on('click.modal', `#${name}`, function (e) {
        if (!$(e.target).closest('.popup__content').length) {
            closeModal(name);
        }
    });

    // Закрытие по клавише Escape
    $(document).on('keydown.modal', function (e) {
        if (e.key === 'Escape') {
            closeModal(name);
        }
    });
}


function closeModal(name) {
    $('.popup#' + name).remove(); // ❗ Тоже удаляем только popup по id
    $('body').removeClass('modal');
    $(document).off('click.modal');
    $(document).off('keydown.modal');
}


// Инициализация страницы
$(document).ready(() => {
    $('#exam_date').focus();

    // Активируем элементы, если они были выбраны ранее
    $('.profile__examiner-item[data-selected="true"]').each(function () {
        $(this).find('.profile__menu-icon').addClass('active');
        $(this).find('.profile__examiner-role').show();
    });

    $('.profile__student-item[data-selected="true"]').each(function () {
        $(this).find('.profile__menu-icon').addClass('active');
    });

    updateExamButtonState();
});

// Обработчики кнопок
$('#create_exam').on('click', (e) => {
    e.preventDefault();
    submitExamForm("/admin/exam/create", false);
});

$('#assign_exam').on('click', (e) => {
    e.preventDefault();
    submitExamForm("/admin/exam/create", true);
});

//---ПОДТВЕРЖДЕНИЕ ЗАЯВКИ НА ЭКЗАМЕН--------------------------------------
//---- Исправленный фрагмент обработки подтверждения и отказа заявления на экзамен ----

// Подтвердить заявление на экзамен
$('#agree_application').on('click', function (e) {
    e.preventDefault();

    const appID = $('body').data('id');

    if (!appID) {
        alert("Ошибка: не удалось получить ID заявления");
        return;
    }

    $.ajax({
        url: '/admin/api/application/approve',
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ id: appID }),
        success: function (response) {
            window.location.href = "/admin/exam/students";
        },
        error: function (xhr) {
            alert("Ошибка подтверждения: " + xhr.responseText);
        }
    });
});

// Открыть форму отказа по кнопке
console.log($('#decline_application_button')); // Проверяем, находит ли jQuery элемент
$('#decline_application_button').on('click', function (e) {
    e.preventDefault();
    console.log("Кнопка отклонения была нажата");
    openModal("decline_application_modal");
});



// Отправка формы отказа
$(document).on('submit', '#decline_application-form', function (e) {
    e.preventDefault();

    if ($('.popup__send').hasClass('popup__send--disabled')) {
        return;
    }

    const reasons = [];
    $('input[name="reason"]:checked').each(function () {
        reasons.push($(this).val());
    });

    const explanation = $('textarea[name="explanation"]').val();
    const appID = $('body').data('id');

    if (!appID) {
        alert("Ошибка: не найден ID заявления");
        return;
    }

    $.ajax({
        type: "POST",
        url: "/admin/api/application/decline", // исправленный маршрут
        contentType: "application/json",
        data: JSON.stringify({
            id: appID,
            reasons: reasons,
            explanation: explanation
        }),
        success: function () {
            window.location.href = "/admin/exam/students";
        },
        error: function (xhr) {
            alert("Ошибка отказа: " + xhr.responseText);
        }
    });
});

$(document).on('click', '.profile__exam-link[data-action]', async function (e) {
    e.preventDefault();

    const $button = $(this);
    const action = $button.data('action');
    const examId = $(this).closest('[data-id]').data('id') || $button.data('id');

    if (!examId) {
        alert('ID экзамена не найден.');
        return;
    }

    try {
        if (action === 'assign') {
            await apiPost('/admin/api/exam/schedule', { id: examId });
            location.reload();
        } else if (action === 'open') {
            await apiPost('/admin/api/exam/set', { id: examId });
            window.location.href = "/admin/exam/show";
        }
    } catch (xhr) {
        showAlert('Ошибка: ' + (xhr.responseText || xhr.statusText), 'error');
        $button.prop('disabled', false).text(action === 'assign' ? 'Назначить' : 'Посмотреть');
    }
});

$(document).on('click', '.profile__exam-link', function (e) {
    e.preventDefault();

    const examID = $(this).closest('[data-id]').data('id'); // получаем ID экзамена из ближайшего родителя

    if (!examID) {
        alert('ID экзамена не найден.');
        return;
    }

    $.ajax({
        type: "POST",
        url: "/admin/api/exam/set",
        contentType: "application/json",
        data: JSON.stringify({ id: examID }),
        success: function (res) {
            if (res.success) {
                window.location.href = "/admin/exam/view";
            } else {
                alert("Ошибка при установке экзамена.");
            }
        },
        error: function (xhr) {
            alert("Ошибка сервера: " + (xhr.responseText || xhr.statusText));
        }
    });
});

$(document).on('input', 'textarea.profile__textarea-auto_expand', function () {
    this.style.height = "auto";
    this.style.height = (this.scrollHeight) + "px";
});

$(document).on("click", ".popup__decline-form", function (e) {
    e.preventDefault();
    $("#reason_decline").addClass("popup__open");
    $("body").addClass("modal");
});

$(document).on("click", ".popup__close", function (e) {
    e.preventDefault();
    $("#reason_decline").removeClass("popup__open");
    $("body").removeClass("modal");
});

$(document).ready(function () {
    setTimeout(refreshAccessToken, 1000); // через 1 секунду
    setInterval(refreshAccessToken, 10 * 60 * 1000);
});



