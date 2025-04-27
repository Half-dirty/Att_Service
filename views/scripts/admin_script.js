// //----GLOBAL FUNCTIONS-----------------------------------------------------------------------
// //--функция для показа алерта
// function showAlert(message, type = "success") {
//     const alertBox = $("#custom-alert");
//     alertBox.removeClass("hidden error show");
//     alertBox.text(message);

//     if (type === "error") {
//         alertBox.addClass("error");
//     }

//     alertBox.addClass("show");

//     // Показываем на 3 секунды, потом скрываем
//     setTimeout(() => {
//         alertBox.removeClass("show");
//     }, 3000);
// }

// //---MODALS AND WORK WITH THEM-----------------------------------------
// const adminModals = {
//     'decline_form':
//         `
//     <div id="decline_form" class="popup">
//         <div class="popup__body">
//             <div class="popup__content">
//                 <div class="popup__header">
//                     <a href="" class="popup__close close-popup">
//                         <span></span>
//                     </a>
//                     <h2 class="popup__title">Укажите причину отказа:</h2>
//                 </div>
//                 <div class="popup__form">
//                     <div class="popup__checker">
//                         <form class="" id="decline-forma" enctype="multipart/form-data" method="POST" action="#">
//                             <div class="popup__list">
//                                 <label class="popup__item">
//                                     <input type="checkbox" name="reason" value="invalid_name"> Неверно указанное ФИО
//                                 </label>
//                                 <label class="popup__item">
//                                     <input type="checkbox" name="reason" value="invalid_contacts"> Неверно указанные контакты
//                                 </label>
//                                 <label class="popup__item">
//                                     <input type="checkbox" name="reason" value="no_documents"> Прикреплены не все документы
//                                 </label>
//                                  <div class="popup__textarea">
//                                     <textarea placeholder="Напишите пояснение" name="explanation"></textarea>
//                                 </div>
//                             </div>
//                             <div class="popup__send popup__send--disabled">
//                                 <button type="submit" class="">Отправить</button>
//                             </div>
//                         </form>
//                     </div>
//                 </div>
//             </div>
//         </div>
//     </div>
//     `,
// }

// function openModal(name) {
//     // Добавляем модальное окно в код
//     $(adminModals[name]).prependTo('.wrapper');

//     const dialog = $('#' + name);

//     // Делаем модальное окно видимым
//     dialog.addClass('popup__open');

//     $('body').addClass('modal');

//     // Закрытие по крестику
//     dialog.find('.close-popup').click(function (e) {
//         e.preventDefault();
//         closeModal(name);
//     });

//     // Закрытие по клавише Esc
//     $(document).keydown(function (e) {
//         if (e.key === 'Escape' && $('body').hasClass('modal')) {
//             closeModal(name);
//         }
//     });

//     // Закрытие по клику на фоне
//     dialog.click(function (e) {
//         if (!$(e.target).closest('.popup__content').length) {
//             closeModal(name);
//         }
//     });
// }

// function closeModal(name) {
//     const dialog = $('#' + name);
//     dialog.removeClass('popup__open');
//     $('body').removeClass('modal');

//     // Удаляем модальное окно из DOM
//     dialog.remove();
// }


// //---USERS PAGES-------------------------------------------------------------------

// //--USER LIST PAGE-----------------------------------

// //--поиск всех людей
// $('#search_all_input').on('input', function (e) {
//     e.preventDefault();

//     let surname = $('#search_all_input').val().trim(); // Получаем введённое значение
//     let content = $('.profile__user_list').empty(); // Очищаем текущий список пользователей

//     // Если поле пустое, показываем весь список
//     if (surname.length === 0) {
//         $.ajax({
//             type: "POST",
//             url: "/admin/search/all",
//             contentType: "application/json",
//             data: JSON.stringify({ surname: '' }),  // Пустой поиск возвращает всех пользователей
//             success: function (res) {
//                 let buf = "";
//                 if (res['success']) {
//                     res['users'].forEach(function (user) {
//                         buf += generateUserHTML(user); // Функция для генерации HTML
//                     });
//                 } else {
//                     buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
//                 }
//                 content.append(buf); // Добавляем пользователей в список
//             }
//         });
//     } else {
//         // Если поле не пустое, выполняем поиск по фамилии
//         $.ajax({
//             type: "POST",
//             url: "/admin/search/all",
//             contentType: "application/json",
//             data: JSON.stringify({ surname: surname }),
//             success: function (res) {
//                 let buf = "";
//                 if (res['success']) {
//                     res['users'].forEach(function (user) {
//                         buf += generateUserHTML(user); // Функция для генерации HTML
//                     });
//                 } else {
//                     buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
//                 }
//                 content.append(buf); // Добавляем пользователей в список
//             }, error: function (xhr, status, error) {
//                 showAlert("Ошибка при сохранении данных!", "error");
//                 console.error('AJAX Error:', status, error);
//             }
//         });
//     }
// });

// $(document).on('click', '.delete-student', function (e) {
//     e.preventDefault();

//     if (!confirm('Вы точно хотите удалить пользователя?')) {
//         return;
//     }

//     $.ajax({
//         type: "POST",
//         url: "/admin/api/student",
//         contentType: "application/json",
//         data: JSON.stringify({ id: id }),
//         success: function () {
//             // Затем удаляем
//             $.ajax({
//                 type: "POST",
//                 url: "/admin/student/delete",
//                 contentType: "application/json",
//                 data: JSON.stringify({}),
//                 success: function () {
//                     location.reload();
//                 },
//                 error: function () {
//                     showAlert("Ошибка при удалении пользователя", "error");
//                 }
//             });
//         },
//         error: function () {
//             showAlert("Ошибка при установке пользователя", "error");
//         }
//     });
// });


// $(document).on('click', '.open-student-profile', function (e) {
//     e.preventDefault();

//     const id = $(this).data('id');
//     const source = $(this).data('source') || "";

//     $.ajax({
//         type: "POST",
//         url: "/admin/api/student",
//         contentType: "application/json",
//         data: JSON.stringify({ id: id, source: source }),
//         success: function () {
//             window.location.href = "/admin/student/profile";
//         },
//         error: function () {
//             showAlert("Ошибка при открытии профиля", "error");
//         }
//     });
// });



// //--фильтр по ролям
// $('#users_role').on('change', function (e) {
//     e.preventDefault();
//     let role = $('#users_role').val();
//     console.log(role);
//     let content = $('.profile__user_list').empty();
//     $.ajax({
//         type: "POST",
//         url: "/admin/select",
//         contentType: "application/json",
//         data: JSON.stringify({ role: role }),
//         success: function (res) {
//             let buf = "";
//             if (res.success) {
//                 for (let i = 0; i < res['users'].length; i++) {
//                     let user = res['users'][i];
//                     buf += `<div class="profile__user" id="${user.id}">
//                             <div class="profile__user-name">
//                                 <h2>
//                                     <img src="${user.avatar}">
//                                     ${user.surname} ${user.name} ${user.lastname}
//                                     <div class="header__role">
//                                         ${user.role === 'admin' ? '<h2 class="header__role-text header__role-text--admin">Администратор</h2>' :
//                             user.role === 'student' ? '<h2 class="header__role-text header__role-text--student">Аттестуемый</h2>' :
//                                 user.role === 'examiner' ? '<h2 class="header__role-text header__role-text--examiner">Экзаменатор</h2>' : ''
//                         }
//                                     </div>
//                                 </h2>
//                             </div>
//                             <div class="profile__user-selector">
//                                 <ul class="profile__user-menu">
//                                     <li><button class="profile__user-link open-student-profile" data-id="${user.id}">Посмотреть аккаунт</button></li>
//                                     <li><button class="profile__user-link delete-student" data-id="${user.id}">Удалить аккаунт</button></li>
//                                 </ul>
//                             </div>
//                         </div>`;
//                 }
//             } else {
//                 buf += '<div class="profile__user">' +
//                     '    <div class="profile__user-name">' +
//                     '       <h2>пользователи не найдены</h2>' +
//                     '    </div>' +
//                     '   <div class="profile__user-selector">' +
//                     '        <div class="profile__menu-icon">' +
//                     '            <span></span>' +
//                     '        </div>' +
//                     '       </div>' +
//                     '</div>'
//             }
//             content.append(buf);
//         }
//     })
// })


// //--USER APPLICATIONS PAGE-----------------------------------

// // Слушаем ввод в поле для поиска в реальном времени
// $('#search_application_input').on('input', function (e) {
//     e.preventDefault();

//     let surname = $('#search_application_input').val().trim(); // Получаем введённое значение
//     let content = $('.profile__user_list').empty(); // Очищаем текущий список пользователей

//     // Если поле пустое, показываем весь список
//     if (surname.length === 0) {
//         $.ajax({
//             type: "POST",
//             url: "/admin/search/application",
//             contentType: "application/json",
//             data: JSON.stringify({ surname: '' }),  // Пустой поиск возвращает всех пользователей
//             success: function (res) {
//                 let buf = "";
//                 if (res['success']) {
//                     res['users'].forEach(function (user) {
//                         buf += generateUserHTML(user); // Функция для генерации HTML
//                     });
//                 } else {
//                     buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
//                 }
//                 content.append(buf); // Добавляем пользователей в список
//             }
//         });
//     } else {
//         // Если поле не пустое, выполняем поиск по фамилии
//         $.ajax({
//             type: "POST",
//             url: "/admin/search/application",
//             contentType: "application/json",
//             data: JSON.stringify({ surname: surname }),
//             success: function (res) {
//                 let buf = "";
//                 if (res['success']) {
//                     res['users'].forEach(function (user) {
//                         buf += generateUserHTML(user); // Функция для генерации HTML
//                     });
//                 } else {
//                     buf += generateUserNotFound(); // Функция для отображения "пользователь не найден"
//                 }
//                 content.append(buf); // Добавляем пользователей в список
//             }, error: function (xhr, status, error) {
//                 showAlert("Ошибка при сохранении данных!", "error");
//                 console.error('AJAX Error:', status, error);
//             }
//         });
//     }
// });

// // Функция для генерации HTML для каждого пользователя
// function generateUserHTML(user) {
//     let roleText = '';
//     switch (user.role) {
//         case "admin":
//             roleText = '<h2 class="header__role-text header__role-text--admin">Администратор</h2>';
//             break;
//         case "student":
//             roleText = '<h2 class="header__role-text header__role-text--student">Аттестуемый</h2>';
//             break;
//         case "examiner":
//             roleText = '<h2 class="header__role-text header__role-text--examiner">Экзаменатор</h2>';
//             break;
//     }

//     return `
//         <div class="profile__user" id="${user.id}">
//             <div class="profile__user-name">
//                 <h2>
//                     <img src="${user.avatar}">
//                     ${user.surname} ${user.name} ${user.lastname}
//                     <div class="header__role">
//                         ${roleText}
//                     </div>
//                 </h2>
//             </div>
//             <div class="profile__user-selector">
//                 <ul class="profile__user-menu">
//                     <li><button class="profile__user-link open-student-profile" data-id="${user.id} data-source="application">Посмотреть аккаунт</button></li>
//                     <li><button class="profile__user-link delete-student" data-id="${user.id}">Удалить аккаунт</button></li>
//                 </ul>
//             </div>
//         </div>
//     `;
// }


// // Функция для вывода, если пользователей не найдено
// function generateUserNotFound() {
//     return '<div class="profile__user">' +
//         '    <div class="profile__user-name">' +
//         '       <h2>Пользователи не найдены</h2>' +
//         '    </div>' +
//         '   <div class="profile__user-selector">' +
//         '        <div class="profile__menu-icon">' +
//         '            <span></span>' +
//         '        </div>' +
//         '       </div>' +
//         '</div>';
// }


// //--фильтр по ролям (хз зачем)
// $('#users_role_application').on('change', function (e) {
//     e.preventDefault();
//     let role = $('#users_role_application').val();
//     console.log(role);
//     let content = $('.profile__user_list').empty();

//     $.ajax({
//         type: "POST",
//         url: "/admin/select/application",
//         contentType: "application/json",
//         data: JSON.stringify({ role: role }),
//         success: function (res) {
//             let buf = "";

//             if (res.success) {
//                 for (let i = 0; i < res.users.length; i++) {
//                     let user = res.users[i];

//                     let roleText = '';
//                     switch (user.role) {
//                         case "admin":
//                             roleText = '<h2 class="header__role-text header__role-text--admin">Администратор</h2>';
//                             break;
//                         case "student":
//                             roleText = '<h2 class="header__role-text header__role-text--student">Аттестуемый</h2>';
//                             break;
//                         case "examiner":
//                             roleText = '<h2 class="header__role-text header__role-text--examiner">Экзаменатор</h2>';
//                             break;
//                     }

//                     buf += `
//                         <div class="profile__user" id="${user.id}">
//                             <div class="profile__user-name">
//                                 <h2>
//                                     <img src="${user.avatar}">
//                                     ${user.surname} ${user.name} ${user.lastname}
//                                     <div class="header__role">
//                                         ${roleText}
//                                     </div>
//                                 </h2>
//                             </div>
//                             <div class="profile__user-selector">
//                                 <ul class="profile__user-menu">
//                                     <li><button class="profile__user-link open-student-profile" data-id="${user.id}" data-source="application">Посмотреть аккаунт</button></li>
//                                     <li><button class="profile__user-link delete-student" data-id="${user.id}">Удалить аккаунт</button></li>
//                                 </ul>
//                             </div>
//                         </div>
//                     `;
//                 }
//             } else {
//                 buf += `
//                     <div class="profile__user">
//                         <div class="profile__user-name">
//                             <h2>Пользователи не найдены</h2>
//                         </div>
//                         <div class="profile__user-selector">
//                             <div class="profile__menu-icon">
//                                 <span></span>
//                             </div>
//                         </div>
//                     </div>
//                 `;
//             }

//             content.append(buf);
//         },
//         error: function () {
//             showAlert("Ошибка при фильтрации по роли", "error");
//         }
//     });
// });



// //--USER SHOW PAGES-----------------------------------

// //--изменение роли пользователя (выпадающий список)
// $('#role-select').on('change', function (e) {
//     e.preventDefault();

//     let role = $(this).val();

//     // удаляем старые классы цвета
//     $(this).removeClass('header__role-select--student header__role-select--examiner');

//     // добавляем новый класс цвета
//     if (role === "examiner") {
//         $(this).addClass('header__role-select--examiner');
//     } else {
//         $(this).addClass('header__role-select--student');
//     }

//     // Просто отправляем роль
//     $.ajax({
//         type: "POST",
//         url: "/admin/change_role", // Без id в query!
//         contentType: "application/json",
//         data: JSON.stringify({ role: role }),
//         success: function (res) {
//             if (res.success) {
//                 showAlert("Роль успешно обновлена!");
//             } else {
//                 showAlert("Ошибка при обновлении роли!", "error");
//             }
//         },
//         error: function (xhr, status, error) {
//             showAlert("Ошибка при обновлении роли!", "error");
//             console.error('AJAX Error:', status, error);
//         }
//     });
// });





// //--подтвердить профиль
// $('#decision-accept').on('click', function (e) {
//     e.preventDefault();
//     $.ajax({
//         type: "POST",
//         url: "/admin/student/confirm",
//         contentType: "application/json",
//         data: JSON.stringify({ confirm: true }),
//         success: function (res) {
//             if (res.success) {
//                 window.location.href = "/admin/user/application";
//             } else {
//                 showAlert("Не удалось подтвердить пользователя", "error");
//             }
//         }
//     });
// });

// // Глобальная функция setupList теперь доступна в любом месте
// function setupList({ selectorIcon, selectorButton, labelOn, labelOff, roleSelectorClass }) {
//     const icons = document.querySelectorAll(selectorIcon);
//     const button = document.querySelector(selectorButton);
//     if (!button) return;

//     function checkAllSelected() {
//         return Array.from(icons).every(icon => icon.classList.contains('active'));
//     }

//     function updateButtonLabel() {
//         const allSelected = checkAllSelected();
//         button.textContent = allSelected ? labelOff : labelOn;
//     }

//     icons.forEach(icon => {
//         icon.addEventListener('click', () => {
//             icon.classList.toggle('active');
//             const roleSelector = icon.closest('.profile__examiner-item')?.querySelector(roleSelectorClass);
//             if (icon.classList.contains('active')) {
//                 roleSelector?.style.setProperty('display', 'block');
//             } else {
//                 roleSelector?.style.setProperty('display', 'none');
//             }
//             updateButtonLabel();
//         });
//     });

//     button.addEventListener('click', function (e) {
//         e.preventDefault();
//         const allSelected = checkAllSelected();
//         const shouldActivate = !allSelected;

//         icons.forEach(icon => {
//             icon.classList.toggle('active', shouldActivate);
//             const roleSelector = icon.closest('.profile__examiner-item')?.querySelector(roleSelectorClass);
//             if (shouldActivate) {
//                 roleSelector?.style.setProperty('display', 'block');
//             } else {
//                 roleSelector?.style.setProperty('display', 'none');
//             }
//         });

//         updateButtonLabel();
//     });
// }

// $('.profile__examiner-item[data-selected="true"]').each(function () {
//     const icon = $(this).find('.profile__menu-icon');
//     const selector = $(this).find('.profile__examiner-role');
//     icon.addClass('active');
//     selector.show(); // отображаем селект роли
// });

// $('.profile__student-item[data-selected="true"]').each(function () {
//     const icon = $(this).find('.profile__menu-icon');
//     icon.addClass('active');
// });


// // Вызов функций setupList после определения
// setupList({
//     selectorIcon: '.profile__examiner-item .profile__menu-icon',
//     selectorButton: '#select-all_examiner',
//     labelOn: 'Выбрать всех',
//     labelOff: 'Убрать всех',
//     roleSelectorClass: '.profile__examiner-role'
// });

// setupList({
//     selectorIcon: '.profile__student-item .profile__menu-icon',
//     selectorButton: '#select-all_student',
//     labelOn: 'Выбрать всех',
//     labelOff: 'Убрать всех',
//     roleSelectorClass: ''
// });

// //--отклонить профиль (открытие модалки)
// $('.profile__decision-decline').on('click', function (e) {
//     e.preventDefault();
//     openModal("decline_form");
// })

// //--функция, для активации кнопки на отправку формы
// $(document).on('change keyup', '#decline_form input[type="checkbox"], #decline_form textarea', function () {
//     const isChecked = $('#decline_form input[type="checkbox"]:checked').length > 0;
//     const isTextEntered = $('#decline_form textarea').val().trim().length > 0;

//     if (isChecked || isTextEntered) {
//         $('.popup__send').removeClass('popup__send--disabled');
//     } else {
//         $('.popup__send').addClass('popup__send--disabled');
//     }
// });

// $(document).on("click", ".profile__exam-link--cancel", function (e) {
//     e.preventDefault();

//     const examId = $(this).closest(".profile__exam").data("id");

//     if (!examId) {
//         alert("ID экзамена не найден");
//         return;
//     }

//     $.ajax({
//         url: "/admin/api/exam/cancel",
//         method: "POST",
//         contentType: "application/json",
//         data: JSON.stringify({ exam_id: examId }),
//         success: function (res) {
//             location.reload();
//         },
//         error: function (xhr) {
//             alert("Ошибка отмены экзамена: " + xhr.responseText);
//         },
//     });
// });


// //--отправка формы отказа
// //--отправка формы отказа
// $(document).on('submit', '#decline-forma', function (e) {
//     e.preventDefault();

//     if ($('.popup__send').hasClass('popup__send--disabled')) {
//         return;
//     }

//     const reasons = [];
//     $('input[name="reason"]:checked').each(function () {
//         reasons.push($(this).val());
//     });

//     const explanation = $('textarea[name="explanation"]').val();
//     const appID = $('body').data('id'); // ВАЖНО!

//     $.ajax({
//         type: "POST",
//         url: "/admin/student/decline",
//         contentType: "application/json",
//         data: JSON.stringify({
//             id: appID,
//             reasons: reasons,
//             explanation: explanation
//         }),
//         success: function () {
//             window.location.href = "/admin/user/application";
//         },
//         error: function (xhr) {
//             alert("Ошибка при отправке: " + xhr.responseText);
//         }
//     });
// });




// //----EXAM PAGES
// //--EXAM CREATE PAGE-----------------------------------
// //--при загрузке страницы, фокус на поле даты экзамена
// $(document).ready(function () {
//     $('#exam_date').focus();
// });

// $(document).ready(function () {
//     $('#exam_code').inputmask('99-99-99', { autoUnmask: true });

//     function setupList({ selectorIcon, selectorButton, labelOn, labelOff, roleSelectorClass }) {
//         const icons = document.querySelectorAll(selectorIcon);
//         const button = document.querySelector(selectorButton);

//         // вспомогательная функция проверки "все выбраны?"
//         function checkAllSelected() {
//             return Array.from(icons).every(icon => icon.classList.contains('active'));
//         }

//         // обновление текста кнопки в зависимости от состояния
//         function updateButtonLabel() {
//             const allSelected = checkAllSelected();
//             button.textContent = allSelected ? labelOff : labelOn;
//         }

//         // обработчик кликов по иконке
//         icons.forEach(icon => {
//             icon.addEventListener('click', () => {
//                 icon.classList.toggle('active');

//                 const roleSelector = icon.closest('.profile__examiner-item')?.querySelector(roleSelectorClass);
//                 if (icon.classList.contains('active')) {
//                     roleSelector?.style.setProperty('display', 'block');
//                 } else {
//                     roleSelector?.style.setProperty('display', 'none');
//                 }

//                 updateButtonLabel(); // обновляем текст кнопки при каждом клике
//             });
//         });

//         // обработчик кнопки "выбрать/убрать всех"
//         button.addEventListener('click', function (e) {
//             e.preventDefault();
//             const allSelected = checkAllSelected();
//             const shouldActivate = !allSelected;

//             icons.forEach(icon => {
//                 icon.classList.toggle('active', shouldActivate);

//                 const roleSelector = icon.closest('.profile__examiner-item')?.querySelector(roleSelectorClass);
//                 if (shouldActivate) {
//                     roleSelector?.style.setProperty('display', 'block');
//                 } else {
//                     roleSelector?.style.setProperty('display', 'none');
//                 }
//             });

//             updateButtonLabel();
//         });
//     }
// });


// $(document).on('click', '.profile__examiner-item .profile__menu-icon', function (e) {
//     e.preventDefault();
//     let iconStatus = $(this).hasClass('active');
//     const roleSelect = $(this).closest('.profile__examiner-item').find('.profile__examiner-role');
//     if (iconStatus) {
//         roleSelect.show();
//     } else {
//         roleSelect.hide();
//     }
// });

// function getUserList(role) {
//     const roles = {
//         'examiner': {
//             icon: '.profile__examiner-item .profile__menu-icon',
//             item: '.profile__examiner-item',
//             roleSelectorClass: '.profile__examiner-role'
//         },
//         'student': {
//             icon: '.profile__student-item .profile__menu-icon',
//             item: '.profile__student-item',
//             roleSelectorClass: ''
//         }
//     }

//     const icons = document.querySelectorAll(roles[role].icon);
//     let users = [];
//     icons.forEach(icon => {
//         if (icon.classList.contains('active')) {
//             const item = icon.closest(roles[role].item);
//             if (item) {
//                 users.push(item.getAttribute('data-id'));
//                 if (role === 'examiner') {
//                     const roleSelect = item.querySelector(roles[role].roleSelectorClass);
//                     if (roleSelect) {
//                         users.push(roleSelect.value); // добавляем роль экзаменатора (не уверен, что работает)
//                     }
//                 }
//             }
//         }
//     });

//     return users;
// }

// //--Отправка формы создания экзамена
// function submitExamForm(link, autoSchedule) {
//     let formData = new FormData();
//     let examiners = getUserList('examiner');
//     let students = getUserList('student');
//     let date = $('#exam_date').val();
//     let commissionStart = $('#commission_start').val();
//     let commissionEnd = $('#commission_end').val();
//     let chairmanId = null;
//     let secretaryId = null;

//     $('.profile__examiner-item .profile__menu-icon.active').each(function () {
//         const item = $(this).closest('.profile__examiner-item');
//         const id = parseInt(item.attr('data-id'));
//         const role = item.find('.profile__examiner-role').val();

//         if (role === 'chair') {
//             chairmanId = id;
//         } else if (role === 'secretary') {
//             secretaryId = id;
//         }
//     });

//     formData.append('examiners', JSON.stringify(examiners));
//     formData.append('students', JSON.stringify(students));
//     formData.append('date', date);
//     formData.append('commission_start', commissionStart);
//     formData.append('commission_end', commissionEnd);
//     formData.append('auto_schedule', autoSchedule ? 'true' : 'false'); // ВАЖНО
//     formData.append('chairman_id', JSON.stringify(chairmanId));
//     formData.append('secretary_id', JSON.stringify(secretaryId));

//     $.ajax({
//         type: "POST",
//         url: link,
//         cache: false,
//         processData: false,
//         contentType: false,
//         data: formData,
//         success: function (res) {
//             if (res.success) {
//                 showAlert("Экзамен сохранён!");
//                 window.location.href = autoSchedule ? "/admin/exam/scheduled" : "/admin/exam/planning";
//             } else {
//                 showAlert("Ошибка при сохранении!", "error");
//             }
//         },
//         error: function (xhr, status, error) {
//             showAlert("Ошибка при сохранении!", "error");
//             console.error('AJAX Error:', status, error);
//         }
//     });
// }


// // Отправка по кнопке
// $('#create_exam').on('click', function (e) {
//     e.preventDefault();
//     submitExamForm("/admin/exam/create", false); // planned
// });

// $('#assign_exam').on('click', function (e) {
//     e.preventDefault();
//     submitExamForm("/admin/exam/create", true); // scheduled
// });

// // Отправка по Enter внутри любых input'ов
// $(document).ready(function () {
//     $(document).on('keydown', 'input, select, textarea', function (e) {
//         if (e.key === 'Enter') {
//             e.preventDefault(); // если не хочешь чтобы он случайно форму "отправил" куда-то

//             const inputs = $('input, select, textarea')
//                 .filter(':visible:not([disabled])');

//             const idx = inputs.index(this);
//             if (idx > -1 && idx + 1 < inputs.length) {
//                 inputs.eq(idx + 1).focus();
//             }
//         }
//     });
// });

// $(document).ready(function () {
//     $('#agree_application').on('click', function (e) {
//         e.preventDefault();

//         const appID = $('body').data('id');

//         if (!appID) {
//             alert("Ошибка: не удалось получить ID заявления");
//             return;
//         }

//         $.ajax({
//             url: '/admin/api/application/approve', // маршрут, обрабатывающий одобрение
//             method: 'POST',
//             contentType: 'application/json',
//             data: JSON.stringify({ id: appID }),
//             success: function (response) {
//                 window.location.href = "/admin/exam/students";
//             },
//             error: function (xhr) {
//                 alert("Ошибка подтверждения: " + xhr.responseText);
//             }
//         });
//     });
//     $('#decline_application').on('click', function (e) {
//         e.preventDefault();
//         openModal("decline_form");
//     });
// });
// $(document).on("click", ".popup__decline-form", function (e) {
//     e.preventDefault();
//     $("#reason_decline").addClass("popup__open");
//     $("body").addClass("modal");
// });

// $(document).on("click", ".popup__close", function (e) {
//     e.preventDefault();
//     $("#reason_decline").removeClass("popup__open");
//     $("body").removeClass("modal");
// });

// $(document).on('click', 'body.exam-planning .profile__exam-link:contains("Назначить")', function (e) {
//     e.preventDefault();

//     // получаем ID экзамена (нужно добавить data-id в HTML-шаблоне!)
//     const examID = $(this).closest('[data-id]').data('id');

//     if (!examID) {
//         alert("Ошибка: не найден ID экзамена");
//         return;
//     }

//     $.ajax({
//         url: '/admin/api/exam/schedule',
//         method: 'POST',
//         contentType: 'application/json',
//         data: JSON.stringify({ id: examID }),
//         success: function (response) {
//             location.reload();
//         },
//         error: function (xhr) {
//             alert("Ошибка назначения: " + xhr.responseText);
//         }
//     });
// });

// $(document).on('click', '.open-exam', function (e) {
//     e.preventDefault();

//     const examID = $(this).data('id');
//     if (!examID) {
//         alert("ID экзамена не найден");
//         return;
//     }

//     $.ajax({
//         type: "POST",
//         url: "/admin/api/exam/set",
//         contentType: "application/json",
//         data: JSON.stringify({ id: examID }),
//         success: function () {
//             window.location.href = "/admin/exam/show";
//         },
//         error: function () {
//             alert("Ошибка открытия экзамена");
//         }
//     });
// });

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
    'decline_application':
        `
    <div id="decline_application" class="popup">
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
                        <form class="" id="decline_application-form" enctype="multipart/form-data" method="POST" action="#">
                            <div class="popup__list">
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="invalid_name"> Неверно указанные данные
                                </label>
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="invalid_contacts"> Прикреплены не соответствующие фото</label>
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
    $.post("/refresh").fail(() => window.location.href = "/");
};
setInterval(refreshAccessToken, 10 * 60 * 1000);


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

$('#agree_application').on('click', function (e) {
    e.preventDefault();
    const appID = $('body').data('id');

    if (!appID) {
        alert("Ошибка: не удалось получить ID заявления");
        return;
    }

    apiPost('/admin/api/application/approve', { id: appID })
        .done(() => window.location.href = "/admin/exam/students")
        .fail(xhr => alert("Ошибка подтверждения: " + xhr.responseText));
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
