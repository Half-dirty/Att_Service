// socket.js

let socket;
let reconnectInterval;
let lastExaminerList = [];


function initSocket(role, id, name, examId) {
    if (!socket || socket.readyState === WebSocket.CLOSED) {
        connectSocket(role, id, name, examId);
    }
}

function connectSocket(role, id, name, examId) {
    socket = new WebSocket(`ws://${location.host}/ws`);

    socket.onopen = () => {
        console.log("✅ WebSocket открыт");

        // Убираем переподключение, если оно запущено
        if (reconnectInterval) {
            clearInterval(reconnectInterval);
            reconnectInterval = null;
        }

        socket.send(JSON.stringify({
            type: "init_user",
            data: { user_id: id, name: name, role: role, exam_id: examId }
        }));
        setInterval(() => {
            if (socket && socket.readyState === WebSocket.OPEN) {
                socket.send(JSON.stringify({ type: "ping" }));
            }
        }, 30000); // Пинг каждые 30 секунд
    };

    socket.onmessage = (event) => {
        if (!event.data) return;

        let message;
        try {
            message = JSON.parse(event.data);
        } catch (e) {
            console.error("Ошибка разбора JSON:", e);
            return;
        }

        switch (message.type) {
            case "examiner_list":
                lastExaminerList = message.data;

                // Если модалка открыта в данный момент
                if ($('#start-exam').hasClass('popup__open')) {
                    openStartExamModal(lastExaminerList);
                }
                break;
            case "start_exam":
                startExam(message.data);
                break;
            case "open_student":
                openStudent(message.data);
                break;
            case "progress_update":
                const studentId = message.data.studentId;
                const progress = message.data.progress;
                const completed = message.data.completed;

                updateStudentProgress(studentId, progress, completed);
                break;
            case "redirect":
                redirect(message.data);
                break;
            case "chairman_status":
                updateChairmanStatus(message.data);
                break;
            case "ping":
                // Ответ на пинг от сервера
                break;
            case "error":
                console.error("Ошибка:", message.data);
                alert("Ошибка: " + message.data);
                break;
            default:
                console.warn("Неизвестный тип сообщения:", message.type);
        }
    };

    socket.onerror = (error) => {
        console.error("Ошибка WebSocket:", error);
    };

    socket.onclose = (event) => {
        console.warn("WebSocket закрыт:", event);

        if (!reconnectInterval) {
            reconnectInterval = setInterval(() => {
                console.log("♻️ Пытаемся переподключиться...");
                connectSocket(role, id, name, examId);
            }, 5000);
        }
    };
}

$(document).on('click', '.student-item', function () {
    const studentId = $(this).data('id'); // id студента
    const examId = $('body').data('exam-id'); // exam_id с бэка

    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
            type: "open_student",
            data: {
                exam_id: examId,
                student_id: studentId
            }
        }));
    }
});

function updateStudentProgress(studentId, progressPercent, completed) {
    const $studentItem = $("#student-" + studentId);
    const $fill = $studentItem.find(".exam__progress-bar-fill");

    // Обновляем ширину прогресс-бара
    $fill.css('width', progressPercent + '%');

    // Меняем цвет прогресс-бара
    $fill.removeClass('active warning danger');
    if (progressPercent < 30) {
        $fill.addClass('danger');
    } else if (progressPercent < 60) {
        $fill.addClass('warning');
    } else {
        $fill.addClass('active');
    }

    // Если студент полностью оценён — блокируем
    if (completed) {
        $studentItem.addClass('completed');
        $studentItem.css('pointer-events', 'none'); // отключить клик
        $studentItem.css('opacity', '0.5'); // визуально затушить
    }

    // Блокируем или разблокируем остальных
    updateStudentAccess();
}

function updateStudentAccess() {
    const $allStudents = $(".exam__item.student-item");
    const $completedStudents = $(".exam__item.student-item.completed");

    // Если есть студент, которого сейчас все ещё не все оценили — остальные блокируем
    const isAnyInProgress = $allStudents.length !== $completedStudents.length;

    $allStudents.each(function () {
        if ($(this).hasClass('completed')) {
            // Уже оценённый студент — заблокирован
            $(this).css('pointer-events', 'none').css('opacity', '0.5');
        } else {
            if (isAnyInProgress) {
                // Есть студент, который еще не оценен полностью → остальные заблокированы
                $(this).css('pointer-events', 'none').css('opacity', '0.5');
            } else {
                // Все свободны для новой оценки
                $(this).css('pointer-events', '').css('opacity', '');
            }
        }
    });
}



// Отображение оценочного листа
function displayGradeSheet(gradeSheet) {
    document.getElementById('protocol_num').value = gradeSheet.protocol;
    document.getElementById('person_fio').value = gradeSheet.name;

    // Заполнение таблицы критериев
    const tbody = document.querySelector('.exam__question-table tbody');
    tbody.innerHTML = '';  // Очистить таблицу

    gradeSheet.criteria.forEach((criterion, index) => {
        const row = document.createElement('tr');
        row.classList.add('exam__question-row');
        row.innerHTML = `
            <td class="exam__question-marks">${index + 1}</td>
            <td class="exam__question-text">${criterion.question}</td>
            <td class="exam__question-marks">
                <input type="radio" name="q${index + 1}" value="0" id="q${index + 1}-0" hidden>
                <div class="exam__radio-label" data-radio-for="q${index + 1}-0"></div>
            </td>
            <td class="exam__question-marks">
                <input type="radio" name="q${index + 1}" value="1" id="q${index + 1}-1" hidden>
                <div class="exam__radio-label" data-radio-for="q${index + 1}-1"></div>
            </td>
            <td class="exam__question-marks">
                <input type="radio" name="q${index + 1}" value="2" id="q${index + 1}-2" hidden>
                <div class="exam__radio-label" data-radio-for="q${index + 1}-2"></div>
            </td>
            <td class="exam__question-marks">
                <input type="radio" name="q${index + 1}" value="3" id="q${index + 1}-3" hidden>
                <div class="exam__radio-label" data-radio-for="q${index + 1}-3"></div>
            </td>
            <td class="exam__question-marks">
                <input type="radio" name="q${index + 1}" value="4" id="q${index + 1}-4" hidden>
                <div class="exam__radio-label" data-radio-for="q${index + 1}-4"></div>
            </td>
            <td class="exam__question-marks">
                <input type="radio" name="q${index + 1}" value="5" id="q${index + 1}-5" hidden>
                <div class="exam__radio-label" data-radio-for="q${index + 1}-5"></div>
            </td>
        `;
        tbody.appendChild(row);
    });
}

document.getElementById('subscribe_button').addEventListener('click', function () {
    const isAbstained = document.getElementById('exam__abstain').checked;
    const scores = [];
    const recommendations = document.getElementById('recomendation').value;
    const qualification = document.getElementById('qualification').value;
    const specialization = document.getElementById('specialization').value;
    const studentId = parseInt(document.body.dataset.studentId, 10); // 🔥 исправлено
    const examId = parseInt(document.body.dataset.examId, 10);        // 🔥 исправлено

    if (!isAbstained) {
        const rows = document.querySelectorAll(".exam__question-row");
        rows.forEach((row, index) => {
            const selected = row.querySelector("input[type='radio']:checked");
            if (selected) {
                scores.push(parseInt(selected.value));
            } else {
                scores.push(null);
            }
        });
    }

    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
            type: "save_grade",
            data: {
                exam_id: examId,
                student_id: studentId,
                scores: isAbstained ? [] : scores,
                qualification: qualification,
                specialization: specialization,
                recommendations: recommendations,
                abstained: isAbstained
            }
        }));
    }
});


function updateChairmanStatus(status) {
    const chairmanStatusText = document.getElementById('chairman_status_text');
    if (!chairmanStatusText) return;

    // Статус председателя не должен меняться на "отсутствует", если он продолжает работать
    if (status === 'present') {
        chairmanStatusText.textContent = "Ожидание действий председателя";
    } else {
        chairmanStatusText.textContent = "Председатель отсутствует";
    }
}

// Открытие модалки с обновлённым списком экзаменаторов
function openStartExamModal(examiners) {
    const list = $('#examiner_list');

    if (list.length === 0) {
        console.warn("Не найден контейнер для списка экзаменаторов");
        return;
    }

    list.empty();

    if (examiners.length === 0) {
        list.append('<p>Пока никто не подключился</p>');
    }

    examiners.forEach(examiner => {
        const statusClass = examiner.status === 'online' ? 'popup__examiner-status--online' : 'popup__examiner-status--offline';
        const statusText = examiner.status === 'online' ? 'Подключён' : 'Не в сети';
        const avatarPath = examiner.avatar || '/static/img/avatar.png';

        const examinerHTML = `
            <div class="popup__examiner-item" data-id="${examiner.id}">
                <div class="popup__examiner-info">
                    <div class="popup__examiner-avatar">
                        <img src="${avatarPath}" alt="Avatar">
                    </div>
                    <div class="popup__examiner-name">
                        <h2>${examiner.name}</h2>
                    </div>
                    <div class="popup__examiner-role">
                        <h2>Экзаменатор</h2>
                    </div>
                </div>
                <div class="popup__examiner-status ${statusClass}">
                    <h2>${statusText}</h2>
                </div>
            </div>`;

        list.append(examinerHTML);
    });
}


function startExam(data) {
    if (data && data.url) {
        window.location.href = data.url;
    } else {
        console.warn("Нет URL для старта экзамена");
    }
}

function openStudent(data) {
    if (data && data.url) {
        window.location.href = data.url;
    } else {
        console.warn("Нет URL для открытия студента");
    }
}



function redirect(data) {
    if (data && data.url) {
        window.location.href = data.url;
    } else {
        console.warn("Нет URL для редиректа");
    }
}
