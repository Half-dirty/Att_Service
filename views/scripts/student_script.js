//----GLOBAL FUNCTIONS-----------------------------------------------------

//---модальные окна и работа с ними

//--модальные окна
const modals = {
    "aplication_form":
        `
    <div id="aplication_form" class="popup">
            <div class="popup__body">
                <div class="popup__content">
                    <div class="popup__header">
                        <a href="" class="popup__close close-popup">
                            <span></span>
                        </a>
                        <h2 class="popup__title">Чтобы подтвердить статус и подать/посмотреть заявки, необходимо
                            отправить данные на проверку</h2>
                    </div>
                    <div class="popup__form">
                        <div class="popup__checker">
                            <h2 class="popup__subtitle">Вы не заполнили следующие поля:</h2>
                            <h2 class="popup__subtitle popup-nonvisible">Вы заполнили все поля. Отправить данные на проверку?</h2>
                            <ul class="popup__list">
                                <li class="popup__item">
                                    пример
                                </li>
                                <li class="popup__item">
                                    пример
                                </li>
                                <li class="popup__item">
                                    пример
                                </li>
                            </ul>
                            <div class="popup__send popup__send--disabled">
                                <button type="button" class="">Отправить</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `,
    "aproove_form":
        `
    <div id="aproove_form" class="popup">
            <div class="popup__body">
                <div class="popup__content">
                    <div class="popup__header">
                        <a href="" class="popup__close close-popup">
                            <span></span>
                        </a>
                        <h2 class="popup__title">Вам отказали в подтверждении профиля</h2>
                    </div>
                    <div class="popup__form">
                        <div class="popup__checker">
                            <ul class="popup__list">
                                <li class="popup__item">
                                    пример
                                </li>
                                <li class="popup__item">
                                    пример
                                </li>
                                <li class="popup__item">
                                    пример
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `,

    "change_photo":
        `
      <div id="change_photo" class="popup">
            <div class="popup__body">
                <div class="popup__content">
                    <div class="popup__header">
                        <a href="" class="popup__close close-popup">
                            <span></span>
                        </a>
                        <h2 class="popup__title">Для изменения фотографии профиля, загрузите фото по следующим правилам:</h2>
                    </div>
                    <div class="popup__form-photo">
                        <div class="popup__photo-rules">
                            <ul class="popup__list">
                                <li class="popup__item">
                                    Размер фотографии должен иметь формат 3х4.
                                </li>
                                <li class="popup__item">
                                    Фон фотографии должен быть однотонным и светлым, белым или светло-серым,
                                    без каких-либо узоров, принтов, логотипов, текстур. Фон не должен иметь теней,
                                    градиентов или других отвлекающих факторов.
                                </li>
                                <li class="popup__item">
                                    Освещение равномерное, без резких теней и бликов. Рекомендуется использовать
                                    мягкое рассеянное освещение, чтобы минимизировать тени на лице. Обе
                                    стороны лица должны быть равномерно освещены.
                                </li>
                                <li class="popup__item">
                                    Нейтральное выражение лица.
                                </li>
                                <li class="popup__item">
                                    Лицо должно быть направлено прямо в камеру, в кадре размещено по центру и
                                    хорошо видно на экране камеры. Голова не должна быть наклонена или повернута
                                    в сторону, взгляд направлен в объектив.
                                </li>
                                <li class="popup__item">
                                    Отсутствие головных уборов.
                                </li>
                                <li class="popup__item">
                                    Никаких лишних предметов в кадре.
                                </li>
                                <li class="popup__item">
                                    Фотография должна быть высокого качества и разрешения, с полной детализацией «картинки».
                                </li>
                            </ul>
                        </div>
                        <div class="popup__add-photo">
                            <div class="profile__scan">
                                <div class="uploadContainer profile__scan--input">
                                    <input type="file" name="scan" id="person_photo" accept="image/*" hidden
                                        data-id="person_img" class="scan__input">
                                    <div id="person_img" class="image-preview img-container">
                                        <img src="#" alt="Выберите изображение">
                                    </div>
                                </div>
                            </div>

                            <div class="popup__send-photo popup__send--disabled">
                                <button type="button" id="change_photo-button" class="">Отправить</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `,
    'reason_decline': `
    <div id="reason_decline" class="popup">
        <div class="popup__body">
            <div class="popup__content">
                <div class="popup__header">
                    <a href="" class="popup__close close-popup">
                        <span></span>
                    </a>
                    <h2 class="popup__title">Причина отказа:</h2>
                </div>
                <div class="popup__form">
                    <div class="popup__checker">
                        <form class="" id="decline-forma" enctype="multipart/form-data" method="POST" action="#">
                            <div class="popup__list">
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="invalid_name" readonly> Неверно указанное ФИО
                                </label>
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="invalid_contacts" readonly> Неверно указанные контакты
                                </label>
                                <label class="popup__item">
                                    <input type="checkbox" name="reason" value="no_documents" readonly> Прикреплены не все документы
                                </label>
                                 <div class="popup__textarea">
                                    <textarea placeholder="Пояснение" name="explanation" readonly></textarea>
                                </div>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </div>
    `
}

//--функции для работы с модалками
function openModal(name) {
    // Добавляем модальное окно в код
    $(modals[name]).prependTo('.wrapper');

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

//--для textarea, автоматически увеличивает размер поля ввода
function adjustHeight(element) {
    element.style.height = "auto"; // Сбрасываем высоту, чтобы корректно измерить новый размер
    element.style.height = (element.scrollHeight) + 5 + "px"; // Устанавливаем высоту по содержимому + 5px padding
}

//--Автоматически подгоняем высоту при загрузке, если есть предзаполненные значения
document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".profile__textarea-auto_expand").forEach(textarea => adjustHeight(textarea));
});


//--модальное окно для отправки на подтверждение профиля
$(".popup__link").click(function (e) {
    e.preventDefault();
    openModal("aplication_form");
    let body = $('.popup__checker').empty();
    let buf = "";

    $.ajax({
        type: "GET",
        url: "/user/data/correct",
        success: function (data) {
            if (data['success']) {
                buf +=
                    '<h2 class="popup__subtitle popup-nonvisible">Вы заполнили все поля. Отправить данные на проверку?</h2>' +
                    '<div class="popup__send">' +
                    '   <button type="button" class="" id="send__for_agree">Отправить</button>'
                '</div>';
            } else {
                let undone_list = data['list'];
                buf +=
                    '<h2 class="popup__subtitle">Вы не заполнили следующие поля:</h2>' +
                    '<ul class="popup__list">';

                for (let i = 0; i < undone_list.length; i++) {
                    buf += '<li class="popup__item">' + undone_list[i] + '</li>';
                }

                buf += '</ul>' +
                    '<div class="popup__send popup__send--disabled">' +
                    '   <button type="button" class="" id="send__for_agree">Отправить</button>' +
                    '</div>';
            }
            body.append(buf);

            $('#send__for_agree').on('click', function (e) {
                e.preventDefault();
                let status = $('#send__for_agree').hasClass('popup__send--disabled');
                if (status) {
                    alert('вы не заполнили все данные');
                    return
                }
                $.ajax({
                    type: "POST",
                    url: "/user/data/aprove",
                    contentType: "application/json",
                    data: JSON.stringify({
                        aprove: true
                    }),
                    success: function (data) {
                        closeModal("aplication_form");
                    }
                })
            })
        }
    })
});

//--WORK WITH PHOTO-----------------------------

//--реагирует на клик по полю инпут, после чего открывает окно добавления фото
$(document).on("click", ".profile__scan--input", function (e) {
    let fileInput = $(this).find("input[type='file']");

    if (fileInput.length) {
        e.stopPropagation();
        fileInput[0].click();
    }
});

//--реагирует на изменения в инпуте, когда добавляется фото, то проверка на то, какой класс есть (связано с модальным окном на изм фото)
$(document).on("change", ".scan__input", function () {
    if ($('.image-preview').hasClass('img-container')) {
        $('.popup__send-photo').removeClass('popup__send--disabled');
    }
    readMultipleFiles(this, $(this).data("id"));
});

//--считывает фотографии из инпута, и предзагружает (визуализирует) их, также изменяет высоту контейнера с фотографиями
function readMultipleFiles(input, imgContainerId) {
    if (input.files && input.files.length > 0) {
        const container = $("#" + imgContainerId);
        const parentDiv = container.closest(".profile__scan--input");
        container.empty(); // Очищаем контейнер перед добавлением новых изображений

        let totalHeight = 0;

        Array.from(input.files).forEach(file => {
            if (file.type.startsWith("image/")) {
                const reader = new FileReader();
                reader.onload = function (e) {
                    const img = $("<img>")
                        .attr("src", e.target.result)
                        .addClass("preview-img");
                    container.append(img); // Добавляем превью в контейнер

                    totalHeight += img[0].naturalHeight || 150; // Добавляем высоту изображения
                    console.log("smth", totalHeight);
                    parentDiv.css('max-height', totalHeight + 20 + 'px'); // Меняем высоту input
                };
                reader.readAsDataURL(file);
            }
        });
    }
}
//--автоподгон фотографии, для страницы main, чтобы "кадрироватьть" фото
jQuery(function ($) {
    function fix_size() {
        var images = $('.img-container img');
        images.each(setsize);

        function setsize() {
            var img = $(this),
                img_dom = img.get(0),
                container = img.parents('.img-container');
            if (img_dom.complete) {
                resize();
            } else img.one('load', resize);

            function resize() {
                if ((container.width() / container.height()) < (img_dom.width / img_dom.height)) {
                    img.width('100%');
                    img.height('auto');
                    return;
                }
                img.height('100%');
                img.width('auto');
            }
        }
    }
    $(window).on('resize', fix_size);
    fix_size();
});

//----MAIN PAGE-------------------------------------------------------------
//--при загрузке страницы, фокус на поле фамилии в ИП
$(document).ready(function () {
    $('#surname_in_ip').focus();
});

//--маски для полей ввода
$('input[type="tel"]').inputmask('+7 (999) 999-99-99', { autoUnmask: true });
$('#passport_serial').inputmask('99-99', { autoUnmask: true });
$('#unit_code').inputmask('999-999', { autoUnmask: true });
$('#passport_num').inputmask('999999', { autoUnmask: true });
$('#snils_num').inputmask('999-999-999 99', { autoUnmask: true });
$('#diplom_num').inputmask('', { autoUnmask: true });

//--открытие модалки на изменение фотопрофиля
$(".user_avatar").on('click', function (e) {
    e.preventDefault();
    openModal('change_photo');
})


//--при клике "отправить" в модалке на изменение фото, идет отправка на сервер этой самой фотографии
$(document).on('click', '#change_photo-button', function (e) {
    e.preventDefault();

    if ($('.popup__send-photo').hasClass('popup__send--disabled')) {
        return; // Не отправляем запрос, если кнопка не активна
    }

    let formData = new FormData();
    formData.append('user_photo', $('#person_photo')[0].files[0]);

    $.ajax({
        type: "POST",
        url: "/user/change/photo",
        cache: false,
        contentType: false,
        processData: false,
        data: formData,
        dataType: 'json',
        success: function (res) {
            if (res['success']) {
                return window.location.href = "/user/profile";
            }
        },
        error: function (xhr, status, error) {
            console.error('AJAX Error:', status, error);
        }
    });
});


/*
СОХРАНЕНИЕ ИНФОРМАЦИИ С ГЛАВНОЙ СТРАНИЦЫ

get {
    error: ""
}

post {
    фамилии:

    surname_in_ip: , - именительный падеж
    surname_in_rp: , - родительный падеж
    surname_in_dp: , - дательный падеж
-------------
    имена:

    name_in_ip: ,
    name_in_rp: ,
    name_in_dp: ,
-------------   
    отчества:

    lastname_in_ip: ,
    lastname_in_rp: ,
    lastname_in_dp: ,   
-------------
    email: ,
    mail: ,
    
    work_phone: ,
    mobile_phone: , 
    
    sex: 
}
'

*/
$('#send__main_page').on('click', function (e) {
    e.preventDefault();
    let surname_in_ip = $('#surname_in_ip').val();
    let surname_in_rp = $('#surname_in_rp').val();
    let surname_in_dp = $('#surname_in_dp').val();
    let name_in_ip = $('#name_in_ip').val();
    let name_in_rp = $('#name_in_rp').val();
    let name_in_dp = $('#name_in_dp').val();
    let lastname_in_ip = $('#lastname_in_ip').val();
    let lastname_in_rp = $('#lastname_in_rp').val();
    let lastname_in_dp = $('#lastname_in_dp').val();
    let email = $('#email').val();
    let mail = $('#mail').val();
    let work_phone = $('#work_phone').val();
    let mobile_phone = $('#mobile_phone').val();
    let sex = $('input[name="sex"]:checked').attr('id');


    if (sex === undefined) {
        sex = '';
    }

    $.ajax({
        type: "POST",
        url: "/user/maindata",
        contentType: "application/json",
        data: JSON.stringify({
            surname_in_ip: surname_in_ip, surname_in_rp: surname_in_rp, surname_in_dp: surname_in_dp,
            name_in_ip: name_in_ip, name_in_rp: name_in_rp, name_in_dp: name_in_dp,
            lastname_in_ip: lastname_in_ip, lastname_in_rp: lastname_in_rp, lastname_in_dp: lastname_in_dp,
            email: email, mail: mail, work_phone: work_phone, mobile_phone: mobile_phone, //нужно ли еще раз запрашивать email..?
            sex: sex
        }),
        success: function (res) {
            if (res.success) {
                showAlert("Данные успешно сохранены!");
            } else {
                showAlert("Ошибка при сохранении данных!", "error");
            }
        },
        error: function (xhr, status, error) {
            showAlert("Ошибка при сохранении данных!", "error");
            console.error('AJAX Error:', status, error);
        }
    })
})

//--при нажатии Enter, триггерим кнопку "Сохранить"
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

//----DOCUMENT PAGE--------------------------------------------------------------------
//--при загрузке страницы, фокус на поле серии паспорта
$(document).ready(function () {
    $('#passport_serial').focus();
});

//--добавление полей для еще одного документа
$(document).ready(function () {
    let documentCounter = 0; // Счётчик добавленных селекторов

    $('.profile__add-placeholder').on('click', function () {
        documentCounter++; // Увеличиваем счётчик

        const template = $(`<div class="profile__selector">
            <div class="profile__selector-header">
                <div class="profile__selector-name--new"></div>
                <button type="button" class="profile__selector--delete-button">Удалить</button>
            </div>
            <div class="profile__form">
                <div class="profile__row">
                    <div class="profile__column">
                        <div class="profile__filler">
                            <label class="profile__label" for="new_doc_num_${documentCounter}">№:</label>
                            <input class="profile__input" type="text" name="new_doc_num_${documentCounter}" id="new_doc_num_${documentCounter}" placeholder="___-___-___ __" required>
                        </div>
                    </div>
                    <div class="profile__column">
                        <label class="profile__label" for="new_doc_text_${documentCounter}">Опишите поля документа в формате<br>(серия: 123322)</label>
                        <textarea class="profile__textarea profile__textarea-auto_expand" rows="1"
                        oninput="adjustHeight(this)" name="new_doc_text_${documentCounter}" id="new_doc_text_${documentCounter}"></textarea>
                    </div>
                </div>
                <div class="profile__row">
                    <div class="profile__column">
                        <div class="profile__scan">
                            <label class="profile__label" for="scan_new_doc_${documentCounter}">Скан:</label>
                            <div class="uploadContainer profile__scan--input">
                                <input type="file" name="scan_new_doc_${documentCounter}" id="scan_new_doc_${documentCounter}" accept="image/*"
                                hidden class="scan__input" data-id="new_doc_img_${documentCounter}" multiple>
                                <div id="new_doc_img_${documentCounter}" class="image-preview">
                                    <img src="#" alt="Выберите изображение">
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>`);

        //запрашиваем список документов, для дальнейшего заполнения (их наименования)
        $.ajax({
            type: "GET",
            url: "/user/document/list",
            success: function (res) {
                let buf = `<input class="profile__selector-name--new-input" type="search" list="new_doc_${documentCounter}" value="Новый документ ${documentCounter}"/>
                <datalist name="new-doc" id="new_doc_${documentCounter}" class="profile__selector-name--new-list">`;

                if (res['success']) {
                    let list = res['documents'];
                    for (let i = 0; i < list.length; i++) {
                        buf += `<option value="${list[i]}"></option>`;
                    }
                }

                buf += `</datalist>`;

                template.find('.profile__selector-name--new').append(buf);
                template.insertBefore($('.profile__add'));
            },
            error: function () {
                showAlert("Ошибка при сохранении данных!", "error");
            }
        });
    });

    // Обработчик для удаления селектора
    $(document).on('click', '.profile__selector--delete-button', function () {
        $(this).closest('.profile__selector').remove();
    });
});


/*
СОХРАНЕНИЕ ИНФОРМАЦИИ СТРАНИЦЫ ДОКУМЕНТОВ

get {
    error: ""
}

post {
    паспорт:

    passport_serial: ,
    unit_code: ,
    passport_num: ,
    passport_date: ,
    passport_issue: ,
    bithday_date: ,
    born_place: ,
    registr_address: ,
    passport_img: ,
-------------
    снилс:

    snils_num: ,
    snils_img: ,
-------------   
    диплом:

    diplom_num: ,
    : ,
    diplom_img: ,   
-------------
    другие:

    new_doc_#цифра: ,
    
    new_doc_num_#цифра: ,
    new_doc_img_#цифра: , 

    и тд
}
'

*/
$('#profile_form').on('submit', function (e) {
    e.preventDefault();
    let formData = new FormData(this);
    for (const [key, value] of formData.entries()) {
        console.log(`${key}: ${value}`);
    }

    $.ajax({
        type: "POST",
        url: "/user/documents/send",
        data: formData,
        processData: false,
        contentType: false,
        success: function (res) {
            if (res.success) {
                showAlert("Данные успешно сохранены!");
            } else {
                showAlert("Ошибка при сохранении данных!", "error");
            }
        },
        error: function (xhr, status, error) {
            showAlert("Ошибка при сохранении данных!", "error");
            console.error('AJAX Error:', status, error);
        }
    })
})

$('.popup__decline-form').on('click', function (e) {
    e.preventDefault();
    openModal('reason_decline');

    let popup__body = $('.popup__list').empty();

    $.ajax({
        type: "GET",
        url: "/user/documents/reason",
        success: function (res) {
            if (res.success) {
                let list = res.list;
                let buf = ``;

                if (list.length > 0) {
                    if (list["invalid_name"]) {
                        buf += `<label class="popup__item">
                            <input type="checkbox" name="reason" value="invalid_name" readonly checked> Неверно указанное ФИО
                        </label>`;
                    }
                    if (list["invalid_contacts"]) {
                        buf += `<label class="popup__item">
                            <input type="checkbox" name="reason" value="invalid_contacts" readonly checked> Неверно указанные контакты
                        </label>`;
                    }
                    if (list["no_documents"]) {
                        buf += `<label class="popup__item">
                            <input type="checkbox" name="reason" value="no_documents" readonly checked> Прикреплены не все документы
                        </label>`;
                    }
                    if (list[explanation]) {
                        buf += `<div class="popup__textarea">
                          <textarea placeholder="Пояснение" name="explanation" readonly>`+ list[exaplanation] + `</textarea>
                        </div>`;
                    }

                    popup__body.append(buf);
                }
            } else {
                showAlert("Ошибка при запросе данных!", "error");
            }
        }, error: function (xhr, status, error) {
            showAlert("Ошибка при запросе данных!", "error");
            console.error('AJAX Error:', status, error);
        }
    })
})

$('#send__application').on('click', function (e) {
    e.preventDefault();

    let native_language = $('#native_language').val();
    let citizenship = $('#citizenship').val();
    let marital_status = $('#marital_status').val();
    let organization = $('#organization').val();
    let job_position = $('#job_position').val();
    let requested_category = $('#requested_category').val();
    let basis_for_attestation = $('#basis_for_attestation').val();
    let existing_category = $('#existing_category').val();
    let existing_category_term = $('#existing_category_term').val();
    let work_experience = $('#work_experience').val();
    let current_position_experience = $('#current_position_experience').val();
    let awards_info = $('#awards_info').val();
    let training_info = $('#training_info').val();
    let memberships = $('#memberships').val();
    let consent = $('#consent').prop('checked');

    if (!consent) {
        showAlert("Нам нужно ваше согласие на обработку данных", "error");
        return;
    }

    $.ajax({
        type: "POST",
        url: "/user/application",
        contentType: "application/json",
        data: JSON.stringify({
            native_language: native_language,
            citizenship: citizenship,
            marital_status: marital_status,
            organization: organization,
            job_position: job_position,
            requested_category: requested_category,
            basis_for_attestation: basis_for_attestation,
            existing_category: existing_category,
            existing_category_term: existing_category_term,
            work_experience: work_experience,
            current_position_experience: current_position_experience,
            awards_info: awards_info,
            training_info: training_info,
            memberships: memberships
        }),
        success: function (res) {
            if (res.success) {
                showAlert("Данные успешно сохранены!");
                window.location.href = "/user/application";
            } else {
                showAlert("Ошибка при сохранении данных!", "error");
            }
        },
        error: function (xhr, status, error) {
            showAlert("Ошибка при сохранении данных!", "error");
            console.error('AJAX Error:', status, error);
        }
    })
})


$('.aproove_form').on('click', function (e) {
    e.preventDefault();
    openModal('aproove_form');

    $.ajax({
        type: "GET",
        url: "/user/decline",
        success: function (res) {
            if (res.success) {
                let list = res.list;
                let buf = ``;
                $('.popup__list').empty();

                if (list["invalid_name"]) {
                    buf += `<label class="popup__item"><input type="checkbox" readonly checked> Неверно указано ФИО</label>`;
                }
                if (list["invalid_contacts"]) {
                    buf += `<label class="popup__item"><input type="checkbox" readonly checked> Неверно указаны контакты</label>`;
                }
                if (list["no_documents"]) {
                    buf += `<label class="popup__item"><input type="checkbox" readonly checked> Не все документы прикреплены</label>`;
                }
                if (list["explanation"]) {
                    buf += `<div class="popup__textarea"><textarea readonly>${list["explanation"]}</textarea></div>`;
                }

                $('.popup__list').append(buf);
            } else {
                showAlert("Ошибка при запросе данных!", "error");
            }
        },
        error: function () {
            showAlert("Ошибка при запросе данных!", "error");
        }
    })
})
