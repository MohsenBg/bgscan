---
title: "bgscan چیست؟"
weight: 1
bookComments: false
---

# bgscan

bgscan یه اسکنر چندپروتکله که با Go نوشته شده و از داخل ترمینال اجرا می‌شه. با استفاده از ICMP، TCP، DNS، HTTP که شامل HTTP/3 هم می‌شه، Xray و تونل‌های DNS مثل DNSTT، VayDNS، Slipstream، MasterDNS، StormDNS و TheFeed می‌تونه آدرس‌های IPv4 و IPv6 رو بررسی کنه.

رابط کاربریش هم کاملاً ترمینالیه و با کیبورد کنترل می‌شه.

با bgscan می‌تونید وضعیت اتصال رو بررسی کنید، اندپوینت‌های مختلف رو با هم مقایسه کنید یا Outboundهای Xray رو تست کنید. نتایج همون لحظه نمایش داده می‌شن و می‌تونید اون‌ها رو به‌صورت CSV ذخیره کنید تا بعداً دوباره ازشون برای اسکن استفاده کنید.

<div align="center">
  <video
    src="https://github.com/user-attachments/assets/4a4c216e-0179-427c-8fe1-89a8fa197730"
    autoplay
    muted
    loop
    playsinline
    controls
    width="100%"
    style="max-width: 900px; border-radius: 12px; border: 1px solid #30363d;"
  ></video>
</div>
