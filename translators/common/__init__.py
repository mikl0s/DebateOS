"""
translators/common — Foundation-neutral shared translator core.

Provides modules shared between all DebateOS translator implementations
(arch, debian, and future foundations):

- contract  : load_resolved_speech, load_opinion_bodies
- manifest  : BuildManifest, derive_source_date_epoch
- firstrun  : render_firstrun_unit, firstrun_unit_name

All modules are foundation-neutral: no Arch-specific or Debian-specific code.
The capability gate (check_capabilities) is NOT here — it is translator-local
because each foundation declares its own supported capability set.
"""
