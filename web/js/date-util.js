// ── 공용 날짜 유틸 ────────────────────────────────────────
// footer.html에서 scripts.js보다 먼저 로드되는 전역 유틸.
// 페이지별 중복 구현 금지 — 날짜 변환/포맷은 여기에만 추가한다.
(function () {
    'use strict';

    // UTC 문자열 → KST(UTC+9) 고정 변환 후 "YYYY. M. D AM/PM h:mm:ss KST" 포맷.
    // 예: 2026. 7. 14 PM 4:12:25 KST / 2026. 12. 3 AM 9:01:08 KST
    // 브라우저 로컬 타임존과 무관하게 항상 KST로 표시한다.
    window.formatKstDate = function (utcString) {
        if (!utcString) return '—';
        const d = new Date(utcString);
        if (isNaN(d)) return String(utcString);
        const k = new Date(d.getTime() + 9 * 60 * 60 * 1000); // KST = UTC+9
        const pad = v => String(v).padStart(2, '0');
        let h = k.getUTCHours();
        const ampm = h < 12 ? 'AM' : 'PM';
        h = h % 12;
        if (h === 0) h = 12;
        return `${k.getUTCFullYear()}. ${k.getUTCMonth() + 1}. ${k.getUTCDate()} ` +
               `${ampm} ${h}:${pad(k.getUTCMinutes())}:${pad(k.getUTCSeconds())} KST`;
    };
})();
