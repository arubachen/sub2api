import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

const wait = (ms: number) => new Promise(resolve => window.setTimeout(resolve, ms))
import HelpTooltip from '@/components/common/HelpTooltip.vue'

function getTooltipElement(): HTMLDivElement {
  const tooltip = document.body.querySelector('[role="tooltip"]')
  if (!(tooltip instanceof HTMLDivElement)) {
    throw new Error('tooltip element not found')
  }
  return tooltip
}

function expectTooltipVisible(tooltip: HTMLElement, visible: boolean) {
  expect(tooltip.style.display === 'none').toBe(!visible)
}

describe('HelpTooltip', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('keeps the existing hover interaction by default', async () => {
    const wrapper = mount(HelpTooltip, {
      attachTo: document.body,
      props: {
        content: 'hover details',
      },
    })

    const trigger = wrapper.get('.group')
    const tooltip = getTooltipElement()

    expectTooltipVisible(tooltip, false)

    await trigger.trigger('mouseenter')
    await nextTick()
    expectTooltipVisible(tooltip, true)

    await trigger.trigger('mouseleave')
    await nextTick()
    await wait(200)
    expectTooltipVisible(tooltip, false)

    wrapper.unmount()
  })

  it('supports click-to-toggle details and closes on outside click', async () => {
    const wrapper = mount(HelpTooltip, {
      attachTo: document.body,
      props: {
        content: 'click details',
        trigger: 'click',
      },
    })

    const trigger = wrapper.get('.group')
    const tooltip = getTooltipElement()

    expectTooltipVisible(tooltip, false)

    await trigger.trigger('click')
    await nextTick()
    expectTooltipVisible(tooltip, true)
    expect(tooltip.textContent).toContain('click details')

    const closeButton = tooltip.querySelector('button[aria-label="Close"]')
    if (!(closeButton instanceof HTMLButtonElement)) {
      throw new Error('close button not found')
    }
    closeButton.click()
    await nextTick()
    await wait(200)
    expectTooltipVisible(tooltip, false)

    await trigger.trigger('click')
    await nextTick()
    expectTooltipVisible(tooltip, true)

    document.body.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await nextTick()
    await wait(200)
    expectTooltipVisible(tooltip, false)

    wrapper.unmount()
  })
})
