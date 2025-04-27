$(document).ready(function () {

    if (!window.socket || socket.readyState === WebSocket.CLOSED) {
        window.socket = new WebSocket(`ws://${location.host}/ws`);

        socket.onopen = () => {
            console.log("✅ WebSocket открыт");
        
            // Получаем данные из <body> (они рендерятся сервером)
            const role = $("body").data("role");
            const id = $("body").data("id");
            const name = $("body").data("name");
            const examId = $("body").data("exam-id");
        
            // Отправляем их сразу после подключения
            socket.send(JSON.stringify({
                type: "init_user",
                data: {
                    user_id: id,
                    name: name,
                    role: role,
                    exam_id: examId
                }
            }));
        };

        socket.onmessage = (event) => {
            if (!event.data) return;

            let message;
            try {
                message = JSON.parse(event.data);
            } catch (e) {
                console.error("Ошибка JSON:", e);
                return;
            }

            switch (message.type) {
                case "connected_list":
                    let html = "";
                    message.data.forEach(name => html += `<li>${name}</li>`);
                    $("#online_examiners").html(html);
                    break;
                case "start_exam":
                    start_exam(message);
                    break;
                case "open_student":
                    open_student(message);
                    break;
                case "redirect":
                    redirect(message);
                    break;
                case "progress_update":
                    progress_update(message);
                    break;
            }
        };

        socket.onerror = (error) => {
            console.error("Ошибка WebSocket: ", error);
        };

        socket.onclose = (event) => {
            console.log("WebSocket закрыт", event);
        };
    }

    $('.exam__list').on('click', '.exam__item', function (e) {
        e.preventDefault();

        let studentId = $(this).data("student-id");
        let currentProgress = parseInt($('#current_progress-' + studentId).val());
        let totalProgress = parseInt($('#total_progress-' + studentId).val());

        if (currentProgress === totalProgress) {
            console.log("Этот студент уже оценен.");
            return;
        }

        console.log("📤 ID выбранного студента:", studentId);

        const selectStudentCommand = {
            type: "select_student",
            data: { studentId: studentId }
        };
        socket.send(JSON.stringify(selectStudentCommand));
    });

    $('#subscribe_button').on('click', function (e) {
        e.preventDefault();
        let studentId = $('.exam__person').data("id");
        let protocol = $('#protocol_num').val();
        let abstain = $('#abstain').prop('checked');
        let score = $('#total').val();
        let recomendation = $('#recomendation').val();
        let qualification = $('#qualification').val();
        let specialization = $('#specialization').val();

        const subscribeCommand = {
            type: "subscribe_document",
            data: {
                studentId, protocol, abstain, score, recomendation, qualification, specialization
            }
        };

        socket.send(JSON.stringify(subscribeCommand));
    });
});

function start_exam(message) {
    const data = message.data;
    window.location.href = data.url;
}

function open_student(message) {
    const data = message.data;
    if (data && data.url) {
        window.location.href = data.url;
    } else {
        console.warn("URL для переадресации отсутствует");
    }
}

function redirect(message) {
    const data = message.data;
    if (data && data.url) {
        window.location.href = data.url;
    } else {
        console.warn("URL для редиректа отсутствует");
    }
}

function progress_update(message) {
    const data = message.data;
    const studentId = data.studentId;
    const currentProgress = data.currentProgress;

    const progressElement = $('#current_progress-' + studentId.toString());
    if (progressElement) {
        progressElement.val(currentProgress);
    }
}

function checkAllStudentsEvaluated() {
    let allEvaluated = true;
    $('.exam__item').each(function () {
        let studentId = $(this).data("student-id");
        let currentProgress = parseInt($('#current_progress-' + studentId).val());
        let totalProgress = parseInt($('#total_progress-' + studentId).val());

        if (currentProgress !== totalProgress) {
            allEvaluated = false;
            return false;
        }
    });

    $('.exam__discuss-button').prop('disabled', !allEvaluated);
}

$('.exam__list').on('input', 'input[type="number"]', function () {
    checkAllStudentsEvaluated();
});


// этот код нужен, чтобы обновлять access_token "в фоне"
const refreshAccessToken = () => {
    $.post("/refresh").fail(() => window.location.href = "/");
};
setInterval(refreshAccessToken, 10 * 60 * 1000);
