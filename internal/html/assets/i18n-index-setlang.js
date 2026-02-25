// Update elements with inline HTML
document.querySelectorAll('[data-i18n-html]').forEach(el => {
  const key = el.dataset.i18nHtml;
  el.innerHTML = t(key);
});

// Update dataflow animation labels
if (window.setDataflowLabels) {
  window.setDataflowLabels({
    yourFile: t('anim_your_file'),
    encrypt: t('anim_encrypt'),
    split: t('anim_split'),
    combine: t('anim_combine'),
    recovered: t('anim_recovered'),
    later: t('anim_later')
  });
}
