$(document).ready(function () {
    //ТЕСТОВЫЕ ДАННЫЕ, ИХ БЫТЬ НЕ ДОЛЖНО, ЭТИ ДАННЫЕ ДОЛЖНЫ ВЫТЯГИВАТЬСЯ ИЗ БД ИЛИ ТОКЕНА!!!
    const role = "admin";  // можно вытаскивать из токена
    const id = "admin1";

    // Только если соединения нет или оно закрыто
    if (!window.socket || socket.readyState === WebSocket.CLOSED) {

        window.socket = new WebSocket("ws://" + location.host + "/ws?role=" + role + "&id=" + id);

        socket.onopen = () => {
            console.log("✅ WebSocket открыт");
        };

        //Слушатель WebSocket сообщений у всех клиентов
        socket.onmessage = (event) => {
            if (!event.data) {
                console.warn("⚠️ Пустое сообщение от WebSocket");
                return;
            }

            let message;
            try {
                message = JSON.parse(event.data);
            } catch (e) {
                console.error("❌ Ошибка парсинга JSON:", e, "Данные:", event.data);
                return;
            }

            switch (message.type) {
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
        // Событие на случай ошибки подключения
        socket.onerror = (error) => {
            console.error("Ошибка WebSocket: ", error);
        };

        // Закрытие соединения
        socket.onclose = (event) => {
            console.log("WebSocket соединение закрыто", event);
        };
    }

    // $('.exam__start-button').on('click', function () {
    //     const command = {
    //         type: "start",
    //         data: {
    //             role: role,
    //             id: id
    //         }
    //     };
    //
    //     if (window.socket && socket.readyState === WebSocket.OPEN) {
    //         socket.send(JSON.stringify(command));
    //         console.log("🚀 Отправлена команда старта экзамена");
    //     } else {
    //         console.warn("⚠️ Сокет не готов, команда не отправлена");
    //     }
    // });

    $('.exam__list').on('click', '.exam__item', function (e) {
        e.preventDefault();

        let studentId = $(this).data("student-id");
        let currentProgress = parseInt($('#current_progress-' + studentId).val());
        let totalProgress = parseInt($('#total_progress-' + studentId).val());

        // Проверяем, если студент уже оценен
        if (currentProgress === totalProgress) {
            console.log("Этот студент уже оценен.");
            return; // Не выбираем студента для оценки, если его прогресс завершен
        }

        console.log("📤 ID выбранного студента:", studentId);

        // Пример отправки команды на сервер
        const selectStudentCommand = {
            type: "select_student",
            data: {
                studentId: studentId
            }
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

        // console.log(studentId, protocol, score, recomendation, qualification, specialization, abstain);

        const subscribeCommand = {
            type: "subscribe_document",
            data: {
                studentId: studentId,
                protocol: protocol,
                abstain: abstain,
                score: score,
                recomendation: recomendation,
                qualification: qualification,
                specialization: specialization
            }
        };

        socket.send(JSON.stringify(subscribeCommand));
    })
})

function start_exam(message) {
    const data = message.data;
    if (message.role === "examiner") {
        // редирект у экзаменаторов
        window.location.href = data.url;
    }
    if (message.role["role"] === "admin") {
        window.location.href = data.url;
    }
}

function open_student(message) {
    const data = message.data;
    const user = data.student[0];

    if (data && data.url) {
        window.location.href = data.url;
    } else {
        console.warn("URL для переадресации отсутствует");
    }
}

// Функция для обработки редиректа
function redirect(message) {
    const data = message.data;
    // console.log(data);
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

    // Обновляем значение прогресса для конкретного студента
    const progressElement = $('#current_progress-' + studentId.toString());
    if (progressElement) {
        progressElement.value = currentProgress;
    }
}

function checkAllStudentsEvaluated() {
    let allEvaluated = true;
    $('.exam__item').each(function() {
        let studentId = $(this).data("student-id");
        let currentProgress = parseInt($('#current_progress-' + studentId).val());
        let totalProgress = parseInt($('#total_progress-' + studentId).val());

        if (currentProgress !== totalProgress) {
            allEvaluated = false; // Если хотя бы один студент не завершил оценку
            return false; // Прерываем цикл
        }
    });

    // Включаем или выключаем кнопку "начать обсуждение"
    if (allEvaluated) {
        $('.exam__discuss-button').prop('disabled', false);
    } else {
        $('.exam__discuss-button').prop('disabled', true);
    }
}

// Проверяем каждый раз, когда изменяется прогресс
$('.exam__list').on('input', 'input[type="number"]', function() {
    checkAllStudentsEvaluated();
});

