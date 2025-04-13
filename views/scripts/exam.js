//--для textarea, автоматически увеличивает размер поля ввода
function adjustHeight(element) {
    element.style.height = "auto"; // Сбрасываем высоту, чтобы корректно измерить новый размер
    element.style.height = (element.scrollHeight) + 5 + "px"; // Устанавливаем высоту по содержимому + 5px padding
}

//--Автоматически подгоняем высоту при загрузке, если есть предзаполненные значения
document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".profile__textarea-auto_expand").forEach(textarea => adjustHeight(textarea));
});


//--маски для полей ввода
$('#protocol_num').inputmask('99-99-99', { autoUnmask: true });

$(document).ready(function () {
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

        // Обновляем поля суммы по столбцам
        columnSums.forEach((sum, index) => {
            $(`#point${index}`).val(sum);
        });

        // Вычисляем средний балл
        let averageScore = totalCount > 0 ? (totalScore / totalCount).toFixed(2) : 0;
        averageScore = Math.round(averageScore * 100) / 100;
        $("#total").val(averageScore);
    }

    $(".exam__question-row input[type='radio']").on("change", updateScores);
    updateScores();

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
            if ($('#recomendation')) {
                $('#recomendation').val('').trigger('input');
            }
        } else {
            $("input[type='text'], input[type='radio'], input[type='date'], select, textarea").not(".exam__buttons-subscribe").prop("disabled", false);
            $("input[name='point'], input[id='total']").val("0");
        }
    });
});
