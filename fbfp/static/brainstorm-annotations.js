const article = document.getElementById('work-text');
const commentDialog = document.getElementById('comment-dialog');
const commentText = document.getElementById('comment-text');
const submitComment = document.getElementById('submit-comment');
const cancelComment = document.getElementById('cancel-comment');
let clickPosition = 0;

article.addEventListener('click', (event) => {
	clickPosition = calculateBytePosition(event);
	commentDialog.style.display = 'block';
});

submitComment.addEventListener('click', () => {
	const comment = commentText.value;
	if (comment) {
		const annotation = {
			position: clickPosition,
			comment: comment
		};
		sendAnnotation(annotation);
	}
	commentDialog.style.display = 'none';
	commentText.value = '';
});

cancelComment.addEventListener('click', () => {
	commentDialog.style.display = 'none';
	commentText.value = '';
});

function calculateBytePosition(event) {
	const range = document.caretRangeFromPoint(event.clientX, event.clientY);
	const preCaretRange = range.cloneRange();
	preCaretRange.selectNodeContents(article);
	preCaretRange.setEnd(range.startContainer, range.startOffset);
	return preCaretRange.toString().length;
}

function sendAnnotation(annotation) {
	const xhr = new XMLHttpRequest();
	xhr.open('POST', '/work/2/annotation/new', true);
	xhr.setRequestHeader('Content-Type', 'application/json;charset=UTF-8');
	xhr.onreadystatechange = function () {
		if (xhr.readyState === 4 && xhr.status === 200) {
			console.log('Annotation sent successfully');
		}
	};
	xhr.send(JSON.stringify(annotation));
}
