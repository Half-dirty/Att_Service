//--для textarea, автоматически увеличивает размер поля ввод
function adjustHeight(el) {
    el.style.height = "auto";
    let maxHeight = 4 * 20; // 4 строки, каждая около 20px высотой
    el.style.height = Math.min(el.scrollHeight, maxHeight) + "px";
    if (el.scrollHeight > maxHeight) {
        el.style.overflowY = "scroll";
    } else {
        el.style.overflowY = "hidden";
    }
}
$(document).ready(function () {
    const role = $("body").data("role");
    const id = $("body").data("id");
    const name = $("body").data("name");
    const examId = $("body").data("exam-id");

    // Проверяем, есть ли данные и если да, инициализируем WebSocket-соединение
    if (role && id && name && examId) {
        initSocket(role, id, name, examId); // ← повторная инициализация при переходе
    }
});


//--Автоматически подгоняем высоту при загрузке, если есть предзаполненные значения
document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".profile__textarea-auto_expand").forEach(textarea => adjustHeight(textarea));
});


$(document).ready(function () {
    //--маски для полей ввода

    function updateScores() {
        let totalScore = 0;

        $(".exam__question-row").each(function () {
            const checkedRadio = $(this).find("input[type='radio']:checked");
            if (checkedRadio.length) {
                totalScore += parseInt(checkedRadio.val(), 10);
            }
        });

        $("#total").val(totalScore);
    }


    // Используем делегирование событий, так как строки создаются динамически:
    $(document).on('click', '.exam__radio-label', function () {
        const radioId = $(this).data('radio-for');
        const radioButton = $('#' + radioId);
    
        if (!radioButton.length) {
            console.warn('Не найдена радиокнопка для', radioId);
            return;
        }
    
        radioButton.prop('checked', true);
    
        // Снимаем выделение у всех соседей
        radioButton.closest('tr').find('.exam__radio-label').removeClass('active');
    
        // Выделяем текущую
        $(this).addClass('active');
    
        updateScores(); // пересчитать баллы
    });
    
    
    

    $(".exam__textarea-auto_expand").on("input", function () {
        adjustHeight(this);
    });

    // Отключение полей при выборе "Воздержаться"
    $(".exam__buttons-abstain input").on("change", function () {
        if ($(this).is(":checked")) {
            $("input[id='recomendation'], input[name='point'], input[id='total'], select, textarea")
                .not(".exam__buttons-subscribe")
                .prop("disabled", true)
                .val("");
            $("input[type='radio']")
                .not(".exam__buttons-subscribe")
                .prop("disabled", true)
                .prop("checked", false);
            $(".exam__radio-label").removeClass('active'); // 🔥 Убираем раскраску
    
            if ($('#recomendation')) {
                $('#recomendation').val('').trigger('input');
            }
        } else {
            $("input[type='text'], input[type='radio'], input[type='date'], select, textarea")
                .not(".exam__buttons-subscribe")
                .prop("disabled", false);
    
            $("input[name='point'], input[id='total']").val("0");
    
            //-- Вот здесь ВАЖНО добавить восстановление активных кнопок:
            $(".exam__radio-label").removeClass('active'); // сброс старого
            $("input[type='radio']:checked").each(function () {
                const id = $(this).attr('id');
                if (id) {
                    $(`.exam__radio-label[data-radio-for="${id}"]`).addClass('active');
                }
            });
        }
    });
    
    updateScores(); // ← чтобы сразу отобразить 0

});
$(document).ready(function () {
    function updateStudentProgress(studentId, progressPercent, grades) {
        const $container = $('#student-' + studentId).find('.exam__progress-bar-container');
        const $fill = $container.find('.exam__progress-bar-fill');
        const $studentBlock = $('#student-' + studentId);

        // Обновление ширины прогресс-бара
        $fill.css('width', progressPercent + '%');

        // Устанавливаем цвет в зависимости от процента прогресса
        let colorClass = '';
        if (progressPercent < 30) {
            colorClass = 'danger';
        } else if (progressPercent < 60) {
            colorClass = 'warning';
        } else {
            colorClass = 'active';
        }
        $fill.removeClass('active warning danger').addClass(colorClass);

        // Обновление содержимого Tooltip
        let $tooltip = $('#tooltip-' + studentId);
        let allGraded = true;
        let tooltipHtml = "";

        grades.forEach(grade => {
            tooltipHtml += `<tr>
                <td>${grade.examiner}</td>
                <td>${grade.status}</td>
                <td>${grade.score !== null ? grade.score : '—'}</td>
            </tr>`;
            if (grade.status !== 'оценил' || grade.score === null) {
                allGraded = false;
            }
        });

        $tooltip.find('tbody').html(tooltipHtml);

        // Подсветка карточки, если все оценки завершены
        if (allGraded) {
            $studentBlock.addClass('completed');
        } else {
            $studentBlock.removeClass('completed');
        }
    }


    // Подключение к WebSocket для получения обновлений
    const socket = new WebSocket(`ws://${location.host}/ws`);
    socket.onmessage = function (event) {
        const data = JSON.parse(event.data);
        if (data.action === 'progress_update') {
            updateStudentProgress(data.studentId, data.progress, data.grades);
        }
    };

    // Отображение Tooltip при наведении
    $(".exam__progress-bar-container").on('mouseenter', function () {
        const studentId = $(this).closest('.exam__item').data('student-id');
        const $tooltip = $('#tooltip-' + studentId);
        const containerRect = this.getBoundingClientRect();

        // Позиционирование Tooltip
        $tooltip.css({
            'display': 'block',
            'left': containerRect.left + 'px',
            'top': (containerRect.bottom + 5) + 'px'
        });

        // Проверяем, помещается ли Tooltip внизу
        const tooltipHeight = $tooltip.outerHeight();
        const spaceBelow = window.innerHeight - containerRect.bottom - 5;
        if (spaceBelow < tooltipHeight) {
            $tooltip.css('top', (containerRect.top - tooltipHeight - 5) + 'px');
        }

        // Проверяем, помещается ли Tooltip по ширине
        const tooltipWidth = $tooltip.outerWidth();
        const spaceRight = window.innerWidth - containerRect.left;
        if (spaceRight < tooltipWidth) {
            $tooltip.css('left', (containerRect.right - tooltipWidth) + 'px');
        }
    }).on('mouseleave', function () {
        const studentId = $(this).closest('.exam__item').data('student-id');
        $('#tooltip-' + studentId).hide();
    });
});


// этот код нужен, чтобы обновлять access_token "в фоне"
const refreshAccessToken = () => {
    $.ajax({
        type: 'POST',
        url: '/refresh',
        xhrFields: { withCredentials: true }, // вот это очень важно!!
        success: () => console.log("Токен успешно обновлен"),
        error: () => window.location.href = "/"
    });
};

setInterval(refreshAccessToken, 10 * 60 * 1000);
