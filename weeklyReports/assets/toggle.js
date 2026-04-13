function toggleChart(id, btn) {
  var elem = document.getElementById(id);

  if (elem.style.display === 'none') {
    elem.style.display = '';
    btn.textContent = 'hide';
  } else {
    elem.style.display = 'none';
    btn.textContent = 'show';
  }
}
