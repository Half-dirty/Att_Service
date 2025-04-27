//--для textarea, автоматически увеличивает размер поля ввода
function adjustHeight(element) {
    element.style.height = "auto"; // Сбрасываем высоту, чтобы корректно измерить новый размер
    element.style.height = (element.scrollHeight) + 5 + "px"; // Устанавливаем высоту по содержимому + 5px padding
}

//--Автоматически подгоняем высоту при загрузке, если есть предзаполненные значения
document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".profile__textarea-auto_expand").forEach(textarea => adjustHeight(textarea));
});


$(document).ready(function () {
    //--маски для полей ввода
    $('#protocol_num').inputmask('99-99-99', {autoUnmask: true});

    $('.exam__radio-label').on('click', function () {
        const radioId = $(this).data('radio-for');
        const radioButton = $('#' + radioId);

        // Убираем выделение со всех радиокнопок
        radioButton.closest('tr').find('input[type="radio"]').prop('checked', false);

        // Устанавливаем флаг "checked" на соответствующую радиокнопку
        radioButton.prop('checked', true);
        console.log(radioButton.prop('checked'), radioButton.attr('id'));


        // Меняем цвет или стиль div, чтобы показать, что он выбран
        $(this).closest('tr').find('div').css('background-color', '#f0f0f0'); // сбрасываем фон
        $(this).css('background-color', '#2f68d5'); // меняем фон выбранного элемента

        updateScores();
    });

    function updateScores() {
        let totalScore = 0;
        let totalCount = 0;
        let columnSums = [0, 0, 0, 0, 0, 0];

        $(".exam__question-row").each(function () {
            let rowChecked = $(this).find("input[type='radio']:checked");
            if (rowChecked.length) {
                let value = parseInt(rowChecked.val());
                totalScore += value;
                totalCount++;
                columnSums[value]++;
            }
        });
        
        let total = 0;
        for (let i = 0; i < columnSums.length; i++) {
            total += i * columnSums[i];
        }

        $("#total").val(totalScore);
    }


    // Ограничение высоты textarea
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

    $(".exam__textarea-auto_expand").on("input", function () {
        adjustHeight(this);
    });

    // Отключение полей при выборе "Воздержаться"
    $(".exam__buttons-abstain input").on("change", function () {
        if ($(this).is(":checked")) {
            $("input[id='recomendation'], input[name='point'], input[id='total'], select, textarea").not(".exam__buttons-subscribe").prop("disabled", true).val("");
            $("input[type='radio']").not(".exam__buttons-subscribe").prop("disabled", true).prop("checked", false);
            const radioIds = $('.exam__radio-label');
            radioIds.each(function () {
                const radioId = $(this).data('radio-for');
                const radioButton = $('#' + radioId);
                radioButton.prop('checked', false);
                $(this).css('background-color', '#f0f0f0'); // сбрасываем фон
            });

            if ($('#recomendation')) {
                $('#recomendation').val('').trigger('input');
            }
        } else {
            $("input[type='text'], input[type='radio'], input[type='date'], select, textarea").not(".exam__buttons-subscribe").prop("disabled", false);
            $("input[name='point'], input[id='total']").val("0");
        }
    });
});


// этот код нужен, чтобы обновлять access_token "в фоне"
const refreshAccessToken = () => {
    $.post("/refresh").fail(() => window.location.href = "/");
};
setInterval(refreshAccessToken, 10 * 60 * 1000);
